package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Student struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StudentSpec   `json:"spec"`
	Status StudentStatus `json:"status"`
}

type StudentSpec struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type StudentStatus struct {
	Phase string `json:"phase"`
}


// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type StudentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Student `json:"items"`
}
