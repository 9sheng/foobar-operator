package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FooBar
type FooBar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FooBarSpec   `json:"spec"`
	Status FooBarStatus `json:"status,omitempty"`
}

// FooBarSpec
type FooBarSpec struct {
	Target  string  `json:"target"`
	Type  *string `json:"type,omitempty"`
}

// FooBarStatus
type FooBarStatus struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	UpdateTime string `json:"updateTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FooBarList
type FooBarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []FooBar `json:"items"`
}

