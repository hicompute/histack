package controller

import (
	"context"

	"github.com/hicompute/histack/api/v1alpha1"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// KubevirtVMIReconciler reconciles a VirtualMachineInstance object
type KubevirtVMIReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubevirt.histack.ir,resources=virtualmachineinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubevirt.histack.ir,resources=virtualmachineinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubevirt.histack.ir,resources=virtualmachineinstances/finalizers,verbs=update
func (r *KubevirtVMIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var vmi kubevirtv1.VirtualMachineInstance
	if err := r.Get(ctx, req.NamespacedName, &vmi); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get VirtualMachine instance.")
		return ctrl.Result{}, err
	}

	interfaces := vmi.Status.Interfaces
	if len(interfaces) == 0 {
		return ctrl.Result{}, nil
	}

	clusterIPList := v1alpha1.ClusterIPList{}
	if err := r.List(ctx, &clusterIPList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.resource", req.Namespace+"/"+req.Name),
		Limit:         -1,
	}); err != nil {
		log.Error(err, "Error on getting cip list")
	}

	releaseIpList := lo.Reduce(clusterIPList.Items, func(result []v1alpha1.ClusterIP, item v1alpha1.ClusterIP, _ int) []v1alpha1.ClusterIP {
		_, ok := lo.Find(interfaces, func(i kubevirtv1.VirtualMachineInstanceNetworkInterface) bool {
			return i.MAC == item.Spec.Mac
		})
		if !ok {
			item.Status.History = append(item.Status.History, v1alpha1.ClusterIPHistory{
				Mac:        item.Spec.Mac,
				ReleasedAt: v1.Now(),
				Interface:  item.Spec.Interface,
				Resource:   item.Spec.Resource,
			})
			item.Spec.Mac = ""
			item.Spec.Interface = ""
			result = append(result, item)
		}
		return result
	}, []v1alpha1.ClusterIP{})

	for i := range releaseIpList {
		if err := r.Update(ctx, &releaseIpList[i]); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubevirtVMIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&kubevirtv1.VirtualMachineInstance{}).
		Named("virtualmachineinstance").
		Complete(r)
}
