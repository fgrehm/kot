package controllers

import (
	"github.com/fgrehm/kot"
	configv1 "github.com/fgrehm/kot/playground/api/v1"
	corev1 "k8s.io/api/core/v1"
)

//+kubebuilder:rbac:groups=core,resources=limitranges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.playground.kot,resources=orgnamespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.playground.kot,resources=orgnamespaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.playground.kot,resources=orgnamespaces/finalizers,verbs=update

var nsReconciler = kot.Reconcile(&kot.One{
	GVK: corev1.SchemeGroupVersion.WithKind("Namespace"),

	Reconcile: kot.SimpleReconcileOne(func(ctx kot.Context, child kot.Object) {
		orgNs := ctx.Resource().(*configv1.OrgNamespace)
		ns := child.(*corev1.Namespace)

		ns.Name = orgNs.Name
		kot.CopyLabels(orgNs, ns)
		kot.CopyAnnotations(orgNs, ns)
	}),

	// Finalizer: kot.WaitForChildrenFinalizer{
	// 	corev1.SchemeGroupVersion.WithKind("Namespace"),
	// },
})

var limitsReconciler = kot.Reconcile(&kot.One{
	GVK: corev1.SchemeGroupVersion.WithKind("LimitRange"),

	If: kot.SimpleIf(func(ctx kot.Context) bool {
		orgNs := ctx.Resource().(*configv1.OrgNamespace)
		return orgNs.Spec.DefaultResources != nil
	}),

	Reconcile: kot.SimpleReconcileOne(func(ctx kot.Context, child kot.Object) {
		orgNs := ctx.Resource().(*configv1.OrgNamespace)
		lr := child.(*corev1.LimitRange)

		lr.Name = "container-defaults"
		lr.Namespace = orgNs.Name
		lr.Spec.Limits = []corev1.LimitRangeItem{}

		defaultRequest := corev1.ResourceList{}
		if req := orgNs.Spec.DefaultResources.Request; req != nil {
			if req.CPU != nil {
				defaultRequest[corev1.ResourceCPU] = *req.CPU
			}
			if req.Memory != nil {
				defaultRequest[corev1.ResourceMemory] = *req.Memory
			}
		}

		defaultLimit := corev1.ResourceList{}
		if limit := orgNs.Spec.DefaultResources.Limit; limit != nil {
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
	// 	orgNs := ctx.Resource().(*configv1.OrgNamespace)
	// 	return orgNs.Spec.DefaultResources != nil
	// }),

	Reconcile: func(ctx kot.Context, children kot.ObjectList) (kot.Result, error) {
		// orgNs := ctx.Resource().(*configv1.OrgNamespace)
		// secrets := children.(*corev1.SecretList)

		// Index list by name

		// For image pull, copy from another NS, accept reference to obj (ex: some-ns/nexus-creds)

		// Set object list afterwards with the values of the secrets map
		return kot.Result{}, nil
	},
})

var statusResolver = kot.ActionFn(func(ctx kot.Context) (kot.Result, error) {
	orgNs := ctx.Resource().(*configv1.OrgNamespace)
	client := kot.ClientDep(ctx)

	nsList := &corev1.NamespaceList{}
	if err := client.List(ctx, nsList, kot.ListChildrenOption(orgNs)); err != nil {
		return kot.Result{}, err
	}
	if len(nsList.Items) != 1 {
		return kot.Result{}, nil
	}

	ns := nsList.Items[0]
	ns.Status.DeepCopyInto(&orgNs.Status.NamespaceStatus)
	return kot.Result{}, nil
})

var OrgNamespaceController = &kot.Controller{
	GVK: configv1.GroupVersion.WithKind("OrgNamespace"),

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
