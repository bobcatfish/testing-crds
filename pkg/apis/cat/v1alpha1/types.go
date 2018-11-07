package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cat is a specification for a Cat resource, which allows the user to create cats inside their k8s cluster
type Cat struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatSpec   `json:"spec"`
	Status CatStatus `json:"status"`
}

// BreedType is the type of breed that the cat is.
type BreedType string

var (
	// BreedTypeMoggie is for a type of cat that is just a moggie
	BreedTypeMoggie BreedType = "moggie"

	// BreedTypeMaineCoone is for big giant maine coons
	BreedTypeMaineCoone BreedType = "maine-coone"

	// BreedTypePersian is for long haired Persain cats
	BreedTypePersian BreedType = "persian"
)

// CatSpec is the spec for a Cat resource
type CatSpec struct {
	Name   string    `json:"name"`
	Phrase string    `json:"phrase"`
	Breed  BreedType `json:"breed"`
}

// CatConditionType is the type for Cat conditions
type CatConditionType string

var (
	// CatConditionTypeNap represents whether or not this cat is taking a nap
	CatConditionTypeNap CatConditionType = "nap"

	// CatConditionTypeHungry represents whether or not this cat is hungry
	CatConditionTypeHungry CatConditionType = "hungry"
)

// CatCondition holds a Condition that the Cat has entered into while being executed
type CatCondition struct {
	Type               CatConditionType       `json:"type"`
	Status             corev1.ConditionStatus `json:"status"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

// CatStatus is the status for a Cat resource
type CatStatus struct {
	Conditions []CatCondition `json:"conditions"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CatList is a list of Cat resources
type CatList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Cat `json:"items"`
}
