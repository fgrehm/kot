package kotclient

import (
	"fmt"
	"sort"
	"strings"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Inspired by https://stackoverflow.com/a/63022947
func ObjectField(obj runtimeclient.Object, field ...string) (interface{}, error) {
	unstructuredObj, err := apiruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	fieldValue, found, err := unstructured.NestedFieldNoCopy(unstructuredObj, field...)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	if !found {
		return unstructured.Unstructured{}, fmt.Errorf("field '%s' not found", strings.Join(field, "."))
	}

	return fieldValue, nil
}

func ExtractList(list runtimeclient.ObjectList) ([]runtimeclient.Object, error) {
	items, err := apimeta.ExtractList(list)
	if err != nil {
		return nil, err
	}

	objs := []runtimeclient.Object{}
	for _, o := range items {
		objs = append(objs, o.(runtimeclient.Object))
	}
	return objs, nil
}

func SetList(list runtimeclient.ObjectList, objs []runtimeclient.Object) error {
	// NOTE: Would be nice to avoid this casting here
	apiObjs := []apiruntime.Object{}
	for _, obj := range objs {
		apiObjs = append(apiObjs, obj)
	}
	return apimeta.SetList(list, apiObjs)
}

func IndexListByUID(list runtimeclient.ObjectList) (map[string]runtimeclient.Object, error) {
	items, err := ExtractList(list)
	if err != nil {
		return nil, err
	}
	idx := map[string]runtimeclient.Object{}
	for _, i := range items {
		uid := string(i.GetUID())
		idx[uid] = i
	}
	return idx, nil
}

func SortByAge(list runtimeclient.ObjectList) error {
	objs, err := ExtractList(list)
	if err != nil {
		return err
	}
	sort.Slice(objs, func(i, j int) bool {
		return objs[i].GetCreationTimestamp().UnixNano() > objs[j].GetCreationTimestamp().UnixNano()
	})

	return SetList(list, objs)
}
