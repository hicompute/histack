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
	"fmt"
	"net"
	"reflect"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/hicompute/histack/api/v1alpha1"
	ipamv1alpha1 "github.com/hicompute/histack/api/v1alpha1"
	netutils "github.com/hicompute/histack/pkg/net_utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterIPPoolReconciler reconciles a ClusterIPPool object
type ClusterIPPoolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ipam.histack.ir,resources=clusterippools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ipam.histack.ir,resources=clusterippools/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ipam.histack.ir,resources=clusterippools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterIPPool object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
func (r *ClusterIPPoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here
	var pool v1alpha1.ClusterIPPool
	if err := r.Get(ctx, req.NamespacedName, &pool); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	_, ipnet, err := net.ParseCIDR(pool.Spec.CIDR)
	if err != nil {
		// Set a degraded condition
		meta.SetStatusCondition(&pool.Status.Conditions, metav1.Condition{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "InvalidCIDR",
			Message: fmt.Sprintf("Invalid CIDR: %v", err),
		})
		_ = r.Status().Update(ctx, &pool)
		return ctrl.Result{}, nil
	}
	totalIPs := netutils.CountUsableIPs(ipnet)
	newStatus := pool.Status.DeepCopy()
	newStatus.TotalIPs = totalIPs.String()
	newStatus.FreeIPs = totalIPs.String()
	if reflect.DeepEqual(&pool.Status, newStatus) {
		return ctrl.Result{}, nil // no changes
	}
	pool.Status = *newStatus
	if err := r.Status().Update(ctx, &pool); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterIPPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ipamv1alpha1.ClusterIPPool{}).
		Named("clusterippool").
		Complete(r)
}
