package ipam

import (
	"context"
	"fmt"
	"time"

	"github.com/hicompute/histack/api/v1alpha1"
	"github.com/hicompute/histack/pkg/k8s"
	netutils "github.com/hicompute/histack/pkg/net_utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
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

func (ipam *IPAM) FindOrCreateClusterIP(r IPAMRequest) (*v1alpha1.ClusterIP, error) {
	ctx := context.Background()
	klog.Infof("%v", r)
	resource := r.Namespace + "/" + r.Name
	var list v1alpha1.ClusterIPList
	if err := ipam.k8sClient.List(ctx, &list); err != nil {
		return nil, err
	}

	if len(list.Items) < 1 {
		return ipam.createClusterIP(r.Interface, r.Mac, r.Family, resource)
	}

	return &list.Items[0], nil
}

func (ipam *IPAM) createClusterIP(iface string, mac *string, ipFamily, resource string) (*v1alpha1.ClusterIP, error) {
	ctx := context.Background()

	ipPool, err := ipam.findEmptyClusterIPPool(ipFamily)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	clusterIP := &v1alpha1.ClusterIP{
		Spec: v1alpha1.ClusterIPSpec{
			ClusterIPPool: ipPool.GetName(),
			Mac:           *mac,
			Interface:     iface,
			Address:       ipAddress,
			Family:        ipFamily,
		},
		Status: v1alpha1.ClusterIPStatus{
			History: []v1alpha1.ClusterIPHistory{{
				AllocatedAt: v1.NewTime(time.Now()),
			}},
		},
	}

	if err := ipam.k8sClient.Create(ctx, clusterIP); err != nil {
		return nil, err
	}

	if err := ipam.k8sClient.Status().Update(ctx, ipPool); err != nil {
		return nil, err
	}

	return clusterIP, nil
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
