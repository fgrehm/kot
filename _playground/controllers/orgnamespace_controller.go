package controllers

import (
	"context"
	"fmt"

	"github.com/fgrehm/kot"
	configv1 "github.com/fgrehm/kot/playground/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:rbac:groups=core,resources=limitranges,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.playground.kot,resources=orgnamespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.playground.kot,resources=orgnamespaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.playground.kot,resources=orgnamespaces/finalizers,verbs=update

var IndexByImportedSecret = kot.Indexer{
	GVK:   configv1.GroupVersion.WithKind("OrgNamespace"),
	Field: ".importedSecret",
	IndexFn: func(resource kot.Object) []string {
		imported := []string{}
		orgNamespace := resource.(*configv1.OrgNamespace)
		for _, secretRef := range orgNamespace.Spec.ImportSecrets {
			imported = append(imported, fmt.Sprintf("%s/%s", secretRef.Namespace, secretRef.Name))
		}
		return imported
	},
}

var secretWatcher = kot.Watch(&kot.ResourceWatcher{
	Watches: &corev1.Secret{},
	When:    kot.ResourceVersionChangedPredicate{},
	Enqueue: func(ctn kot.Container, obj kot.Object) ([]kot.ReconcileRequest, error) {
		client := kot.ClientDep(ctn)
		list := &configv1.OrgNamespaceList{}
		secret := obj.(*corev1.Secret)

		ref := fmt.Sprintf("%s/%s", secret.Namespace, secret.Name)
		filter := kot.MatchingFields{".importedSecret": ref}

		ctx := context.Background()
		if err := client.List(ctx, list, filter); err != nil {
			return nil, err
		}

		reqs := make([]kot.ReconcileRequest, len(list.Items))
		for i, item := range list.Items {
			reqs[i].Name = item.Name
			reqs[i].Namespace = item.Namespace
		}
		return reqs, nil
	},
})

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
		client := kot.ClientDep(ctx)
		orgNamespace := ctx.Resource().(*configv1.OrgNamespace)
		secretsList := children.(*corev1.SecretList)
		idxSecrets := map[string]*corev1.Secret{}

		for i := range secretsList.Items {
			secret := &secretsList.Items[i]
			idxSecrets[secret.Name] = secret
		}

		secrets := []corev1.Secret{}
		for _, secretRef := range orgNamespace.Spec.ImportSecrets {
			sourceSecret := &corev1.Secret{}
			secretKey := kot.ClientKey{Namespace: secretRef.Namespace, Name: secretRef.Name}
			if err := client.Get(ctx, secretKey, sourceSecret); err != nil {
				return kot.Result{}, err
			}

			if existingSecret, ok := idxSecrets[secretRef.Name]; ok {
				existingSecret.Data = sourceSecret.Data
				secrets = append(secrets, *existingSecret)
			} else {
				secrets = append(secrets, corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      secretRef.Name,
						Namespace: orgNamespace.Name,
					},
					Data: sourceSecret.Data,
				})
			}
		}

		secretsList.Items = secrets
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

	Watchers: kot.Watchers{
		secretWatcher,
	},

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
