package testctrls

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fgrehm/kot"
	testapi "github.com/fgrehm/kot/internal/testapi/v1"
	corev1 "k8s.io/api/core/v1"
)

var StaticValue = "0123456789"

var nsWatcher = kot.Watch(&kot.ResourceWatcher{
	Watches: &corev1.Namespace{},
	When:    kot.ResourceVersionChangedPredicate{},
	Enqueue: func(ctn kot.Container, obj kot.Object) ([]kot.ReconcileRequest, error) {
		client := kot.ClientDep(ctn)
		list := &testapi.SimpleCRDList{}

		ctx := context.Background()
		ns := obj.(*corev1.Namespace)
		if err := client.List(ctx, list, kot.InNamespace(ns.Name)); err != nil {
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

var cmReconciler = kot.Reconcile(&kot.One{
	GVK: corev1.SchemeGroupVersion.WithKind("ConfigMap"),

	Reconcile: func(ctx kot.Context, child kot.Object) (kot.Result, error) {
		simpleCRD := ctx.Resource().(*testapi.SimpleCRD)
		cm := child.(*corev1.ConfigMap)

		cm.Name = simpleCRD.Name
		cm.Namespace = simpleCRD.Namespace

		cmValue := simpleCRD.Spec.ConfigMapValue
		if cmValue != nil {
			if *cmValue == "boom" {
				return kot.Result{}, errors.New("boom!!!!")
			}
			if *cmValue == "gone" {
				return kot.Result{}, nil
			}

			cm.Data = map[string]string{"value": *cmValue}
		}

		return kot.Result{}, nil
	},
})

var saReconciler = kot.Reconcile(&kot.One{
	GVK: corev1.SchemeGroupVersion.WithKind("ServiceAccount"),

	If: kot.SimpleIf(func(ctx kot.Context) bool {
		simpleCRD := ctx.Resource().(*testapi.SimpleCRD)
		cmVal := simpleCRD.Spec.ConfigMapValue
		return cmVal == nil || *cmVal != "skip-sa"
	}),

	Reconcile: kot.SimpleReconcileOne(func(ctx kot.Context, child kot.Object) {
		simpleCRD := ctx.Resource().(*testapi.SimpleCRD)
		sa := child.(*corev1.ServiceAccount)

		sa.Name = simpleCRD.Name
		sa.Namespace = simpleCRD.Namespace
	}),
})

var secretsReconciler = kot.Reconcile(&kot.List{
	GVK: corev1.SchemeGroupVersion.WithKind("Secret"),

	Reconcile: kot.SimpleReconcileList(func(ctx kot.Context, list kot.ObjectList) {
		simpleCRD := ctx.Resource().(*testapi.SimpleCRD)
		secrets := list.(*corev1.SecretList)

		if len(secrets.Items) == 2 {
			return
		}

		secret := &corev1.Secret{}
		secret.GenerateName = fmt.Sprintf("%s-", simpleCRD.Name)
		secret.Namespace = simpleCRD.Namespace

		secrets.Items = append(secrets.Items, *secret)
	}),
})

var countReconciler = kot.Reconcile(&kot.Custom{
	Name: "count-changes",

	Reconcile: kot.SimpleAction(func(ctx kot.Context) {
		simpleCRD := ctx.Resource().(*testapi.SimpleCRD)

		cmVal := simpleCRD.Spec.ConfigMapValue
		if cmVal == nil {
			return
		}

		statusVal := simpleCRD.Status.KnownConfigMapValue
		if statusVal != nil && *statusVal == *cmVal {
			return
		}
		Counter.Increment()
	}),
})

var statusResolver = kot.ActionFn(func(ctx kot.Context) (kot.Result, error) {
	simpleCRD := ctx.Resource().(*testapi.SimpleCRD)

	simpleCRD.Status.Finalizing = simpleCRD.DeletionTimestamp != nil
	simpleCRD.Status.StaticValue = &StaticValue
	simpleCRD.Status.KnownConfigMapValue = simpleCRD.Spec.ConfigMapValue
	simpleCRD.Status.KnownSecretValue = simpleCRD.Spec.SecretValue

	client := kot.ClientDep(ctx)
	ns := &corev1.Namespace{}
	if err := client.Get(ctx, kot.ClientKey{Name: simpleCRD.Namespace}, ns); err != nil {
		return kot.Result{}, err
	}

	simpleCRD.Status.NamespaceAnnotation = ""
	if ns.Annotations != nil {
		simpleCRD.Status.NamespaceAnnotation = ns.Annotations["misc"]
	}

	return kot.Result{}, nil
})

var delayFinalizer = &kot.SimpleFinalizer{
	EnabledFn: func(ctx kot.Context) (bool, error) {
		return kot.HasAnnotation(ctx.Resource(), "delay"), nil
	},

	FinalizeFn: func(ctx kot.Context) (bool, kot.Result, error) {
		simpleCRD := ctx.Resource().(*testapi.SimpleCRD)

		delay := kot.GetAnnotation(simpleCRD, "delay")
		d, err := time.ParseDuration(delay)
		if err != nil {
			return false, kot.Result{}, err
		}
		removeAfter := simpleCRD.DeletionTimestamp.Time.Add(d)

		now := time.Now()
		if now.After(removeAfter) {
			return true, kot.Result{}, nil
		}

		diff := removeAfter.Sub(now)
		return false, kot.Result{RequeueAfter: diff + time.Second}, nil
	},
}

var SimpleCRDController = &kot.Controller{
	GVK: testapi.GroupVersion.WithKind("SimpleCRD"),

	Watchers: kot.Watchers{
		nsWatcher,
	},

	Reconcilers: kot.Reconcilers{
		cmReconciler,
		saReconciler,
		secretsReconciler,
		countReconciler,
	},

	StatusResolvers: kot.StatusResolvers{
		statusResolver,
	},

	Finalizers: kot.Finalizers{
		delayFinalizer,
	},
}
