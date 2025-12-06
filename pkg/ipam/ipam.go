package ipam

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hicompute/histack/api/v1alpha1"
	"github.com/hicompute/histack/pkg/k8s"
	netutils "github.com/hicompute/histack/pkg/net_utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IPAM struct {
	k8sClient client.Client
}

func New() *IPAM {
	k8sClient, err := k8s.NewClient()
	if err != nil {
		klog.Fatalf("error on creating k8s client: %v", err)
	}
	return &IPAM{
		k8sClient: k8sClient,
	}
}

func (ipam *IPAM) FindOrCreateClusterIP(r IPAMRequest) (*v1alpha1.ClusterIP, *v1alpha1.ClusterIPPool, error) {
	ctx := context.Background()
	var pod corev1.Pod

	if err := ipam.k8sClient.Get(ctx, types.NamespacedName{Namespace: r.Namespace, Name: r.Name}, &pod); err != nil {
		return nil, nil, err
	}
	kubevirtVM := pod.Labels["vm.kubevirt.io/name"]
	resource := r.Namespace + "/"
	// var vmIface kubevirtv1.VirtualMachineInstanceNetworkInterface
	if kubevirtVM != "" {
		resource += kubevirtVM
	} else {
		resource += r.Name
	}

	var list v1alpha1.ClusterIPList
	if err := ipam.k8sClient.List(ctx, &list, &client.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector("spec.family", r.Family),
			fields.OneTermEqualSelector("spec.containerInterface", r.Interface),
			fields.OneTermEqualSelector("spec.resource", resource),
		),
		Limit: 1000,
	}); err != nil {
		return nil, nil, err
	}
	macPrefix := os.Getenv("MAC_PREFIX")
	if macPrefix == "" {
		macPrefix = "02"
	}

	mac := netutils.GenerateVethMAC(resource, macPrefix)
	if len(list.Items) < 1 {
		return ipam.createClusterIP(r.Interface, &mac, r.Family, resource)
	}
	var ipPool v1alpha1.ClusterIPPool
	if err := ipam.k8sClient.Get(ctx, types.NamespacedName{Name: list.Items[0].Spec.ClusterIPPool}, &ipPool); err != nil {
		return nil, nil, err
	}
	return &list.Items[0], &ipPool, nil
}

func (ipam *IPAM) createClusterIP(iface string, mac *string, ipFamily, resource string) (*v1alpha1.ClusterIP, *v1alpha1.ClusterIPPool, error) {
	ctx := context.Background()

	ipPool, err := ipam.findEmptyClusterIPPool(ipFamily)
	if err != nil {
		return nil, nil, err
	}

	var idx uint64

	if len(ipPool.Status.ReleasedIndexes) > 0 {
		idx = ipPool.Status.ReleasedIndexes[0]
		ipPool.Status.ReleasedIndexes = ipPool.Status.ReleasedIndexes[1:]
		ipPool.Status.FreeIPs--
		ipPool.Status.AllocatedIPs++
	} else {
		idx = ipPool.Status.NextIndex
		ipPool.Status.FreeIPs--
		ipPool.Status.AllocatedIPs++
		if ipPool.Status.FreeIPs > 0 {
			ipPool.Status.NextIndex++
		}
	}

	ipAddress, err := netutils.PickIPFromCIDRindex(ipPool.Spec.CIDR, idx)
	if err != nil {
		return nil, nil, err
	}

	clusterIP := &v1alpha1.ClusterIP{
		ObjectMeta: v1.ObjectMeta{
			Name: strings.Replace(resource, "/", "-", -1) + "-" + iface,
		},
		Spec: v1alpha1.ClusterIPSpec{
			ClusterIPPool: ipPool.GetName(),
			Mac:           *mac,
			Interface:     iface,
			Address:       ipAddress,
			Family:        ipFamily,
			Resource:      resource,
		},
		Status: v1alpha1.ClusterIPStatus{
			History: []v1alpha1.ClusterIPHistory{{
				AllocatedAt: v1.NewTime(time.Now()),
			}},
		},
	}

	if err := ipam.k8sClient.Create(ctx, clusterIP); err != nil {
		klog.Errorf("the error on create clusterIP: %v", err)
		return nil, nil, err
	}

	if err := ipam.k8sClient.Status().Update(ctx, ipPool); err != nil {
		return nil, nil, err
	}

	return clusterIP, ipPool, nil
}

func (ipam *IPAM) findEmptyClusterIPPool(ipFamily string) (*v1alpha1.ClusterIPPool, error) {
	ctx := context.Background()
	var list v1alpha1.ClusterIPPoolList

	err := ipam.k8sClient.List(ctx, &list, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.ipFamily", ipFamily),
		Limit:         10000,
	})
	if err != nil {
		return nil, err
	}
	for _, pool := range list.Items {
		if pool.Status.FreeIPs > 0 {
			return &pool, nil
		}
	}
	return nil, fmt.Errorf("no free %s pool found", ipFamily)
}

func (ipam *IPAM) FindClusterIPbyFamilyandMAC(mac, family string) (*v1alpha1.ClusterIP, error) {
	var list v1alpha1.ClusterIPList
	if err := ipam.k8sClient.List(context.Background(), &list, &client.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector("spec.family", family),
			fields.OneTermEqualSelector("spec.mac", mac),
		),
		Limit: 1,
	}); err != nil {
		return nil, err
	}
	if len(list.Items) > 0 {
		return &list.Items[0], nil
	}
	return nil, fmt.Errorf("not found")
}

func (ipam *IPAM) FindClusterIPPoolByName(name string) (*v1alpha1.ClusterIPPool, error) {
	var cipp v1alpha1.ClusterIPPool
	if err := ipam.k8sClient.Get(context.Background(), client.ObjectKey{Name: name}, &cipp); err != nil {
		return nil, err
	}
	return &cipp, nil
}
