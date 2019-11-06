package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Website struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec WebsiteSpec `json:"spec"`
	Status WebsiteStatus `json:"status"`
}

type WebsiteSpec struct {
    GitRepo string `json:"gitRepo"`
    // TargetDeployment string `json:"targetDeployment"`
    // MinReplicas      int    `json:"minReplicas"`
    // MaxReplicas      int    `json:"maxReplicas"`
    // MetricType       string `json:"metricType"`
    // Step             int    `json:"step"`
    // ScaleUp          int    `json:"scaleUp"`
    // ScaleDown        int    `json:"scaleDown"`
}

type WebsiteStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type WebsiteList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`

    Items []Website `json:"items"`
}