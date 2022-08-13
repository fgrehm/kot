package kotclient

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Key = types.NamespacedName
type GVK = schema.GroupVersionKind
type GV = schema.GroupVersion
type GR = schema.GroupResource

var (
	IsNotFound      = apierrors.IsNotFound
	NewNotFound     = apierrors.NewNotFound
	IgnoreNotFound  = runtimeclient.IgnoreNotFound
	RetryOnConflict = retry.RetryOnConflict
	DefaultRetry    = retry.DefaultRetry
)
