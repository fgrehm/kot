/*
Copyright 2022.

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

package controllers

import (
	"github.com/fgrehm/kot"
	richnsv1 "github.com/fgrehm/kot/richns/api/v1"
	corev1 "k8s.io/api/core/v1"
)

//+kubebuilder:rbac:groups=richns.examples.kot,resources=richnamespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=richns.examples.kot,resources=richnamespaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=richns.examples.kot,resources=richnamespaces/finalizers,verbs=update

var nsReconciler = kot.Reconcile(&kot.One{
	GVK: corev1.SchemeGroupVersion.WithKind("Namespace"),

	Reconcile: kot.SimpleReconcileOne(func(ctx kot.Context, child kot.Object) {
		rns := ctx.Resource().(*richnsv1.RichNamespace)
		ns := child.(*corev1.Namespace)

		ns.Name = rns.Name
	}),
})

var limitsReconciler = kot.Reconcile(&kot.One{
	GVK: corev1.SchemeGroupVersion.WithKind("LimitRange"),

	If: kot.SimpleIf(func(ctx kot.Context) bool {
		rns := ctx.Resource().(*richnsv1.RichNamespace)
		return rns.Spec.DefaultResources != nil
	}),

	Reconcile: kot.SimpleReconcileOne(func(ctx kot.Context, child kot.Object) {
		rns := ctx.Resource().(*richnsv1.RichNamespace)
		lr := child.(*corev1.LimitRange)

		lr.Name = "container-defaults"
		lr.Namespace = rns.Name
		lr.Spec.Limits = []corev1.LimitRangeItem{}

		defaultRequest := corev1.ResourceList{}
		if req := rns.Spec.DefaultResources.Request; req != nil {
			if req.CPU != nil {
				defaultRequest[corev1.ResourceCPU] = *req.CPU
			}
			if req.Memory != nil {
				defaultRequest[corev1.ResourceMemory] = *req.Memory
			}
		}

		defaultLimit := corev1.ResourceList{}
		if limit := rns.Spec.DefaultResources.Limit; limit != nil {
			if limit.CPU != nil {
				defaultLimit[corev1.ResourceCPU] = *limit.CPU
			}
			if limit.Memory != nil {
				defaultLimit[corev1.ResourceMemory] = *limit.Memory
			}
		}

		lr.Spec.Limits = []corev1.LimitRangeItem{{
			Type:           corev1.LimitTypeContainer,
			DefaultRequest: defaultRequest,
			Default:        defaultLimit,
		}}
	}),
})

var secretsReconciler = kot.Reconcile(&kot.List{
	GVK: corev1.SchemeGroupVersion.WithKind("Secret"),

	// If: kot.SimpleIf(func(ctx kot.Context) bool {
	// 	rns := ctx.Resource().(*richnsv1.RichNamespace)
	// 	return rns.Spec.DefaultResources != nil
	// }),

	Reconcile: func(ctx kot.Context, children kot.ObjectList) (kot.Result, error) {
		// rns := ctx.Resource().(*richnsv1.RichNamespace)
		// secrets := children.(*corev1.SecretList)

		// Index list by name

		// For image pull, copy from another NS, accept reference to obj (ex: some-ns/nexus-creds)

		// Set object list afterwards with the values of the secrets map
		return kot.Result{}, nil
	},
})

var statusResolver = kot.SimpleAction(func(ctx kot.Context) {
	// rns := ctx.Resource().(*testapi.RichNamespace)
	// rns.Status.Foo = ...

	// TODO: - Copy phase from ns resource
	//       - If pull secret can't be found, set some condition to false
})

var RichNamespaceController = &kot.Controller{
	GVK: richnsv1.GroupVersion.WithKind("RichNamespace"),

	Reconcilers: kot.Reconcilers{
		nsReconciler,
		limitsReconciler,
		secretsReconciler,
		// saReconciler, // configure default SA with image pull secrets
	},

	StatusResolvers: kot.StatusResolvers{
		statusResolver,
	},
}
