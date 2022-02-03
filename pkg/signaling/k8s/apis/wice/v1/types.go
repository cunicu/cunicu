package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"riasc.eu/wice/pkg/pb"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SignalingEnvelope struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	pb.SignalingEnvelope
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SignalingEnvelopeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SignalingEnvelope `json:"items"`
}
