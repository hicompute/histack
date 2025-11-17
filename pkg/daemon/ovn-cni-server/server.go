package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	types100 "github.com/containernetworking/cni/pkg/types/100"

	"github.com/containernetworking/cni/pkg/version"
	cniTypes "github.com/hicompute/histack/pkg/daemon/ovn-cni-server/types"
	histack_ipam "github.com/hicompute/histack/pkg/ipam"
	netUtils "github.com/hicompute/histack/pkg/net_utils"
	"github.com/hicompute/histack/pkg/ovn"
	"github.com/hicompute/histack/pkg/ovs"
	"k8s.io/klog/v2"
)

type CNIServer struct {
	socketPath string
	listener   net.Listener
	ovsAgent   ovs.OvsAgent
	ovnAgent   ovn.OVNagent
	ipam       histack_ipam.IPAM
}

func Start(socketPath string) error {
	// Cleanup existing socket
	os.RemoveAll(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to create socket: %v", err)
	}

	ovsAgent, err := ovs.CreateOVSagent()
	if err != nil {
		return fmt.Errorf("failed to create ovs agent: %v", err)
	}

	ovnAgent, err := ovn.CreateOVNagent("tcp:192.168.12.177:6641")
	if err != nil {
		return fmt.Errorf("failed to create ovs agent: %v", err)
	}

	ipam := histack_ipam.New()

	cniServer := &CNIServer{
		socketPath: socketPath,
		listener:   listener,
		ovsAgent:   *ovsAgent,
		ovnAgent:   *ovnAgent,
		ipam:       *ipam,
	}

	cniServer.run()
	return nil
}

func (s *CNIServer) run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			klog.Errorf("Failed to accept CNI connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}

}

func (s *CNIServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	var request cniTypes.CNIRequest
	if err := json.NewDecoder(conn).Decode(&request); err != nil {
		klog.Errorf("Failed to accept CNI connection: %v", err)
		return
	}

	var response cniTypes.CNIResponse
	switch request.Cmd {
	case "Add":
		response = s.handleAdd(request.CmdArgs)
	case "Del":
		response = s.handleDel(request.CmdArgs)
	// case "CHECK":
	// 	response = s.handleCheck(request)
	default:
		response = cniTypes.CNIResponse{Error: "Unknown command"}
	}

	json.NewEncoder(conn).Encode(response)
}

func (s *CNIServer) handleAdd(req skel.CmdArgs) cniTypes.CNIResponse {

	k8sArgs := cniTypes.CniKubeArgs{}
	if err := types.LoadArgs(req.Args, &k8sArgs); err != nil {
		klog.Infof("error loading args: %v", err)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	K8S_POD_NAMESPACE := string(k8sArgs.K8S_POD_NAMESPACE)
	K8S_POD_NAME := string(k8sArgs.K8S_POD_NAME)
	clusterIP, clusterIPPool, err := s.ipam.FindOrCreateClusterIP(histack_ipam.IPAMRequest{
		Interface: req.IfName,
		Namespace: K8S_POD_NAMESPACE,
		Name:      K8S_POD_NAME,
		Family:    "v4",
	})

	if err != nil {
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}

	hostIface, contIface, err := netUtils.SetupVeth(req.Netns, req.IfName, clusterIP.Spec.Mac, 1500)
	if err != nil {
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	klog.Info(hostIface.Mac, ",", contIface.Mac)

	ifaceId := K8S_POD_NAMESPACE + "_" + K8S_POD_NAME + "_" + req.IfName

	if err = s.ovsAgent.AddPort("br-int", hostIface.Name, "system", ifaceId); err != nil {
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}

	if err := s.ovnAgent.CreateLogicalPort("public", ifaceId, contIface.Mac); err != nil {
		klog.Errorf("error on creating ovn logical port %s on ls %s: %v", ifaceId, "public", err)
		_ = s.ovsAgent.DelPort("br-int", ifaceId)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	result := current.Result{
		CNIVersion: version.Current(),
		Interfaces: []*current.Interface{contIface},
	}
	_, ipNet, err := net.ParseCIDR(clusterIPPool.Spec.CIDR)
	if req.IfName == "eth0" {
		result.IPs = []*current.IPConfig{
			{
				Interface: types100.Int(0),
				Address:   net.IPNet{IP: net.ParseIP(clusterIP.Spec.Address), Mask: net.IPMask(ipNet.Mask)},
				Gateway:   net.IP(clusterIPPool.Spec.Gateway),
			},
		}
	}
	return cniTypes.CNIResponse{
		Result: result,
		Error:  "",
	}
}

func (s *CNIServer) handleDel(req skel.CmdArgs) cniTypes.CNIResponse {
	k8sArgs := cniTypes.CniKubeArgs{}
	if err := types.LoadArgs(req.Args, &k8sArgs); err != nil {
		klog.Infof("error loading args: %v", err)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	K8S_POD_NAMESPACE := string(k8sArgs.K8S_POD_NAMESPACE)
	K8S_POD_NAME := string(k8sArgs.K8S_POD_NAME)
	ifaceId := K8S_POD_NAMESPACE + "_" + K8S_POD_NAME + "_" + req.IfName
	if err := s.ovsAgent.DelPort("br-int", ifaceId); err != nil {
		klog.Errorf("%v", err)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	if err := s.ovnAgent.DeleteLogicalPort("public", ifaceId); err != nil {
		klog.Errorf("%v", err)
		return cniTypes.CNIResponse{
			Error: err.Error(),
		}
	}
	return cniTypes.CNIResponse{}
}
