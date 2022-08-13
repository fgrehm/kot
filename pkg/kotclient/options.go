package kotclient

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func ForegroundDeletion() runtimeclient.DeleteOption {
	return runtimeclient.PropagationPolicy("Foreground")
}

type MatchingFields = runtimeclient.MatchingFields

func InNamespace(name string) runtimeclient.InNamespace {
	return runtimeclient.InNamespace(name)
}

// NOTE: This will panic if an invalid label is provided, if you have tests then you'll hopefully catch them early :)
func LabeledWith(name, value string) *runtimeclient.ListOptions {
	return &runtimeclient.ListOptions{
		LabelSelector: labelSelector(labelRequirementEq(name, value)),
	}
}

// NOTE: This will panic if an invalid label is provided, if you have tests then you'll hopefully catch them early :)
func WithLabels(lbls map[string]string) *runtimeclient.ListOptions {
	requirements := []*labels.Requirement{}

	for name, value := range lbls {
		requirements = append(requirements, labelRequirementEq(name, value))
	}

	return &runtimeclient.ListOptions{
		LabelSelector: labelSelector(requirements...),
	}
}

func labelSelector(reqs ...*labels.Requirement) labels.Selector {
	sel := labels.NewSelector()
	for _, req := range reqs {
		sel = sel.Add(*req)
	}
	return sel
}

// NOTE: This will panic if an invalid label is provided, if you have tests then you'll hopefully catch them early :)
func labelRequirementEq(label string, values ...string) *labels.Requirement {
	req, err := labels.NewRequirement(label, selection.Equals, values)
	if err != nil {
		panic(err)
	}
	return req
}

// selection.DoesNotExist
// selection.Equals
// selection.DoubleEquals
// selection.In
// selection.NotEquals
// selection.NotIn
// selection.Exists
// selection.GreaterThan
// selection.LessThan
