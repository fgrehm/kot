package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// SimpleCRD is a CRD we use on tests to validate the behavior of our internal
// reconciler. This should be kept simple with only primitive values on spec /
// status since the deepcopy code is manually maintained and it's going to be
// easier that way
type SimpleCRD struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimpleCRDSpec   `json:"spec,omitempty"`
	Status SimpleCRDStatus `json:"status,omitempty"`
}

type SimpleCRDSpec struct {
	ReferencedMap  *string `json:"referencedMap"`
	ConfigMapValue *string `json:"configMapValue"`
	SecretValue    *string `json:"secretValue"`
}

type SimpleCRDStatus struct {
	StaticValue         *string `json:"staticValue"`
	ReferencedValue     *string `json:"referencedValue"`
	KnownConfigMapValue *string `json:"knownConfigMapValue"`
	KnownSecretValue    *string `json:"knownSecretValue"`
	NamespaceAnnotation string  `json:"namespaceAnnotation"`
	Finalizing          bool    `json:"finalizing"`
}

type SimpleCRDList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SimpleCRD `json:"items"`
}

func (in *SimpleCRD) DeepCopyInto(out *SimpleCRD) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

func (in *SimpleCRD) DeepCopy() *SimpleCRD {
	if in == nil {
		return nil
	}
	out := new(SimpleCRD)
	in.DeepCopyInto(out)
	return out
}

func (in *SimpleCRD) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *SimpleCRDList) DeepCopyInto(out *SimpleCRDList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SimpleCRD, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SimpleCRDList.
func (in *SimpleCRDList) DeepCopy() *SimpleCRDList {
	if in == nil {
		return nil
	}
	out := new(SimpleCRDList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SimpleCRDList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SimpleCRDSpec) DeepCopyInto(out *SimpleCRDSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SimpleCRDSpec.
func (in *SimpleCRDSpec) DeepCopy() *SimpleCRDSpec {
	if in == nil {
		return nil
	}
	out := new(SimpleCRDSpec)
	in.DeepCopyInto(out)
	return out
}

func init() {
	SchemeBuilder.Register(&SimpleCRD{}, &SimpleCRDList{})
}
