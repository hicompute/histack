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

package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	ipamv1alpha1 "github.com/hicompute/histack/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var clusterippoollog = logf.Log.WithName("clusterippool-resource")

// SetupClusterIPPoolWebhookWithManager registers the webhook for ClusterIPPool in the manager.
func SetupClusterIPPoolWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&ipamv1alpha1.ClusterIPPool{}).
		WithValidator(&ClusterIPPoolCustomValidator{}).
		WithDefaulter(&ClusterIPPoolCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-ipam-histack-ir-v1alpha1-clusterippool,mutating=true,failurePolicy=fail,sideEffects=None,groups=ipam.histack.ir,resources=clusterippools,verbs=create;update,versions=v1alpha1,name=mclusterippool-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterIPPoolCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind ClusterIPPool when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type ClusterIPPoolCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &ClusterIPPoolCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind ClusterIPPool.
func (d *ClusterIPPoolCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	clusterippool, ok := obj.(*ipamv1alpha1.ClusterIPPool)

	if !ok {
		return fmt.Errorf("expected an ClusterIPPool object but got %T", obj)
	}
	clusterippoollog.Info("Defaulting for ClusterIPPool", "name", clusterippool.GetName())

	// TODO(user): fill in your defaulting logic.

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-ipam-histack-ir-v1alpha1-clusterippool,mutating=false,failurePolicy=fail,sideEffects=None,groups=ipam.histack.ir,resources=clusterippools,verbs=create;update,versions=v1alpha1,name=vclusterippool-v1alpha1.kb.io,admissionReviewVersions=v1

// ClusterIPPoolCustomValidator struct is responsible for validating the ClusterIPPool resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ClusterIPPoolCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &ClusterIPPoolCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ClusterIPPool.
func (v *ClusterIPPoolCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	clusterippool, ok := obj.(*ipamv1alpha1.ClusterIPPool)
	if !ok {
		return nil, fmt.Errorf("expected a ClusterIPPool object but got %T", obj)
	}
	clusterippoollog.Info("Validation for ClusterIPPool upon creation", "name", clusterippool.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ClusterIPPool.
func (v *ClusterIPPoolCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	clusterippool, ok := newObj.(*ipamv1alpha1.ClusterIPPool)
	if !ok {
		return nil, fmt.Errorf("expected a ClusterIPPool object for the newObj but got %T", newObj)
	}
	clusterippoollog.Info("Validation for ClusterIPPool upon update", "name", clusterippool.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ClusterIPPool.
func (v *ClusterIPPoolCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	clusterippool, ok := obj.(*ipamv1alpha1.ClusterIPPool)
	if !ok {
		return nil, fmt.Errorf("expected a ClusterIPPool object but got %T", obj)
	}
	clusterippoollog.Info("Validation for ClusterIPPool upon deletion", "name", clusterippool.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
