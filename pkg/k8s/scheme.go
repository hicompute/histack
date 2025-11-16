package k8s

import (
	v1alpha1 "github.com/hicompute/histack/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

var Scheme = runtime.NewScheme()

func init() {
	_ = v1alpha1.AddToScheme(Scheme)
	// add other APIs here (Core, Apps, CRDs, etc.)
}
