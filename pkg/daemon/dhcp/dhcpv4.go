package dhcp

import (
	"net"
	"os"
	"strings"

	histack_ipam "github.com/hicompute/histack/pkg/ipam"
	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"k8s.io/klog/v2"
)

type HDHCPV4 struct {
	ipam   histack_ipam.IPAM
	server server4.Server
}

func Start() error {
	laddr := &net.UDPAddr{IP: net.IPv4zero, Port: 67}

	ipam := histack_ipam.New()
	h := &HDHCPV4{ipam: *ipam}

	srv, err := server4.NewServer("br-ext", laddr, h.handler, server4.WithDebugLogger())
	if err != nil {
		return err
	}
	return srv.Serve()
}

func (hd4 *HDHCPV4) handler(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {
	macPrefix := os.Getenv("MAC_PREFIX")
	if macPrefix == "" {
		macPrefix = "02"
	}
	mac := req.ClientHWAddr.String()
	if !strings.HasPrefix(mac, macPrefix) {
		return
	}

	// Log basic info
	klog.Infof("=== DHCP %s from %s ===", req.MessageType(), req.ClientHWAddr)
	klog.Infof("XID: %s CIAddr: %s SIAddr: %s", req.TransactionID, req.ClientIPAddr, req.ServerIPAddr)

	switch req.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		hd4.handleDiscover(conn, peer, req)
	case dhcpv4.MessageTypeRequest:
		hd4.handleRequest(conn, peer, req)
	default:
		klog.Infof("Ignoring DHCP message: %s", req.MessageType())
	}
}

// ------------------------------------------------------------
//  SHARED HELPERS (DRY)
// ------------------------------------------------------------

// buildReply creates a reply packet (Offer or ACK)
func buildReply(req *dhcpv4.DHCPv4, msgType dhcpv4.MessageType) (*dhcpv4.DHCPv4, error) {
	reply, err := dhcpv4.NewReplyFromRequest(req)
	if err != nil {
		return nil, err
	}

	reply.UpdateOption(dhcpv4.OptMessageType(msgType))

	return reply, nil
}

// applyCommonOptions sets mask, server ID, routes, etc.
func applyCommonOptions(pkt *dhcpv4.DHCPv4) {
	// Subnet mask /24
	pkt.UpdateOption(dhcpv4.OptSubnetMask(net.CIDRMask(24, 32)))

	// Server IP
	pkt.ServerIPAddr = net.ParseIP("172.16.17.66")
	pkt.UpdateOption(dhcpv4.OptServerIdentifier(net.ParseIP("172.16.17.66")))
	// Classless static routes
	route1 := &dhcpv4.Route{
		Dest: &net.IPNet{
			IP:   net.ParseIP("172.16.22.1"),
			Mask: net.CIDRMask(32, 32),
		},
		Router: net.ParseIP("0.0.0.0"),
	}

	route2 := &dhcpv4.Route{
		Dest: &net.IPNet{
			IP:   net.ParseIP("0.0.0.0"),
			Mask: net.CIDRMask(0, 32),
		},
		Router: net.ParseIP("172.16.22.1"),
	}

	pkt.UpdateOption(dhcpv4.OptClasslessStaticRoute(route1, route2))
}

// sendPacket writes the DHCP packet
func sendPacket(conn net.PacketConn, peer net.Addr, pkt *dhcpv4.DHCPv4) {
	_, err := conn.WriteTo(pkt.ToBytes(), peer)
	if err != nil {
		klog.Errorf("send failed: %v", err)
	}
}

// ------------------------------------------------------------
//              OFFER HANDLER
// ------------------------------------------------------------

func (hd4 *HDHCPV4) handleDiscover(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {
	ip, err := hd4.ipam.FindClusterIPbyFamilyandMAC(req.ClientHWAddr.String(), "v4")
	if err != nil {
		klog.Errorf("IPAM lookup failed: %v", err)
		return
	}

	offer, err := buildReply(req, dhcpv4.MessageTypeOffer)
	if err != nil {
		klog.Errorf("Offer reply build failed: %v", err)
		return
	}

	offer.YourIPAddr = net.ParseIP(ip.Spec.Address)

	applyCommonOptions(offer)

	klog.Infof("Sending OFFER → %s", offer.YourIPAddr)

	sendPacket(conn, peer, offer)
}

// ------------------------------------------------------------
//              REQUEST HANDLER
// ------------------------------------------------------------

func (hd4 *HDHCPV4) handleRequest(conn net.PacketConn, peer net.Addr, req *dhcpv4.DHCPv4) {

	ip := req.RequestedIPAddress()
	if ip == nil {
		klog.Errorf("Client REQUEST had no RequestedIPAddress")
		return
	}

	ack, err := buildReply(req, dhcpv4.MessageTypeAck)
	if err != nil {
		klog.Errorf("Ack reply build failed: %v", err)
		return
	}

	ack.YourIPAddr = ip

	applyCommonOptions(ack)

	klog.Infof("Sending ACK → %s", ack.YourIPAddr)

	sendPacket(conn, peer, ack)
}
