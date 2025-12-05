/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/hicompute/histack/api/v1alpha1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// KubeVirtReconciler reconciles a KubeVirt object
type KubeVirtReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Add RBAC permissions for VirtualMachines
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch
// +kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines/status,verbs=get
func (r *KubeVirtReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	log.Info("Reconciling VirtualMachine", "namespace", req.Namespace, "name", req.Name)

	var vm kubevirtv1.VirtualMachine
	if err := r.Get(ctx, req.NamespacedName, &vm); err != nil {
		if errors.IsNotFound(err) {
			log.Info("VirtualMachine deleted, updating associated ClusterIP",
				"namespace", req.Namespace, "name", req.Name)

		}
	}
	clusterIPList := v1alpha1.ClusterIPList{}
	if err := r.List(ctx, &clusterIPList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.resource", req.Namespace+"/"+req.Name),
		Limit:         1000,
	}); err != nil {
		log.Error(err, "Failed to list ClusterIPs for VM", "vm", req.Name)
		return ctrl.Result{}, err
	}
	log.Info("Reconciling VirtualMachine", "namespace", clusterIPList)
	return ctrl.Result{}, nil
}

func (r *KubeVirtReconciler) SetupWithManager(mgr ctrl.Manager) error {
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
