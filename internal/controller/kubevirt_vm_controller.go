package controller

import (
	"context"
	"fmt"

	"github.com/jaswdr/faker/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hicompute/histack/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// KubeVirtVMReconciler reconciles a KubeVirt object
type KubeVirtVMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Add RBAC permissions for VirtualMachines
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines/status,verbs=get
func (r *KubeVirtVMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var vm kubevirtv1.VirtualMachine
	if err := r.Get(ctx, req.NamespacedName, &vm); err != nil {
		if errors.IsNotFound(err) {
			return r.handleVMDeletion(ctx, req.Namespace, req.Name, metav1.Now())
		}
		log.Error(err, "Failed to get VirtualMachine")
		return ctrl.Result{}, err
	}

	if vm.DeletionTimestamp != nil {
		return r.handleVMDeletion(ctx, req.Namespace, req.Name, *vm.DeletionTimestamp)
	}

	if vm.Status.ObservedGeneration != vm.GetGeneration() {
		// return r.handleVMUpdate(ctx, vm)
		return ctrl.Result{}, nil
	}

	// handle vm creation.
	fake := faker.New()
	vmCredentialsSecret := &corev1.Secret{}
	vmCredentialsSecret.Name = fmt.Sprintf("%s-credentials", vm.Name)
	vmCredentialsSecret.Data = map[string][]byte{
		"username": []byte(fake.Internet().User()),
		"password": []byte(fake.Internet().Password()),
	}

	err := r.Client.Create(ctx, vmCredentialsSecret)
	if (err != nil) && (!errors.IsAlreadyExists(err)) {
		log.Error(err, "Failed to create VM credentials secret", "vm", vm.Name)
		return ctrl.Result{}, err
	}

	vm.Spec.Template.Spec.AccessCredentials = []kubevirtv1.AccessCredential{
		{
			UserPassword: &kubevirtv1.UserPasswordAccessCredential{
				Source: kubevirtv1.UserPasswordAccessCredentialSource{
					Secret: &kubevirtv1.AccessCredentialSecretSource{
						SecretName: vmCredentialsSecret.Name,
					},
				},
				PropagationMethod: kubevirtv1.UserPasswordAccessCredentialPropagationMethod{
					QemuGuestAgent: &kubevirtv1.QemuGuestAgentUserPasswordAccessCredentialPropagation{},
				},
			},
		},
	}
	if err := r.Update(ctx, &vm); err != nil {
		log.Error(err, "Failed to update VM with access credentials", "vm", vm.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KubeVirtVMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1alpha1.ClusterIP{}, "spec.resource", func(rawObj client.Object) []string {
		cip := rawObj.(*v1alpha1.ClusterIP)
		return []string{cip.Spec.Resource}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubevirtv1.VirtualMachine{}).
		// Watches(&kubevirtv1.VirtualMachine{}, &handler.EnqueueRequestForObject{}).
		Named("virtualmachine").
		Complete(r)
}

func (r *KubeVirtVMReconciler) handleVMDeletion(ctx context.Context, namespace, vmName string, deletedAt v1.Time) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	vmCredentials := corev1.Secret{}
	vmCredentials.Name = fmt.Sprintf("%s-credentials", vmName)
	vmCredentials.Namespace = namespace

	if err := r.Client.Delete(ctx, &vmCredentials); err != nil {
		log.Error(err, "Failed to delete vm credentials", "vm", vmName)
	}

	// clientset.CoreV1().Secrets(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})

	clusterIPList := v1alpha1.ClusterIPList{}
	if err := r.List(ctx, &clusterIPList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.resource", namespace+"/"+vmName),
		Limit:         1000,
	}); err != nil {
		log.Error(err, "Failed to list ClusterIPs for VM", "vm", vmName)
		return ctrl.Result{}, err
	}
	log.Info("Reconciling VirtualMachine", "namespace", clusterIPList)

	for _, clusterIP := range clusterIPList.Items {
		updatedClusterIP := clusterIP.DeepCopy()
		updatedClusterIP.Spec.Mac = ""
		updatedClusterIP.Spec.Interface = ""
		updatedClusterIP.Spec.Resource = ""
		updatedClusterIP.Status.History = append(updatedClusterIP.Status.History,
			v1alpha1.ClusterIPHistory{
				Mac:         clusterIP.Spec.Mac,
				Resource:    namespace + "/" + vmName,
				AllocatedAt: *clusterIP.Status.History[len(clusterIP.Status.History)-1].AllocatedAt.DeepCopy(),
				ReleasedAt:  deletedAt,
				Interface:   clusterIP.Spec.Interface,
			},
		)
		if err := r.Status().Update(ctx, updatedClusterIP); err != nil {
			log.Error(err, "Failed to update ClusterIP status", "clusterip", updatedClusterIP.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// func (r *KubeVirtVMReconciler) handleVMUpdate(ctx context.Context, vm kubevirtv1.VirtualMachine) (ctrl.Result, error) {

// }
