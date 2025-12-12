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

	"github.com/hicompute/histack/api/v1alpha1"
	ipamv1alpha1 "github.com/hicompute/histack/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// ClusterIPReconciler reconciles a ClusterIP object
type ClusterIPReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ipam.histack.ir,resources=clusterips,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipam.histack.ir,resources=clusterips/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ipam.histack.ir,resources=clusterips/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterIP object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
func (r *ClusterIPReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	var clusterIP v1alpha1.ClusterIP
	if err := r.Client.Get(ctx, client.ObjectKey{Name: req.Name}, &clusterIP); err != nil {
		return ctrl.Result{}, err
	}

	if clusterIP.Spec.Mac == "" {
		// handler release
		var clusterIPPool v1alpha1.ClusterIPPool
		if err := r.Client.Get(ctx, client.ObjectKey{Name: clusterIP.Spec.ClusterIPPool}, &clusterIPPool); err != nil {
			return ctrl.Result{}, err
		}
		pool := clusterIPPool.DeepCopy()
		pool.Status.ReleasedClusterIPs = append(pool.Status.ReleasedClusterIPs, clusterIP.GetName())
		pool.Status.FreeIPs++
		pool.Status.AllocatedIPs--
		if err := r.Client.Status().Update(ctx, pool); err != nil {
			log.Error(err, "update cluster ip pool failed.")
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterIPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipamv1alpha1.ClusterIP{}).
		Named("clusterip").
		Complete(r)
}
