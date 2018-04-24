package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Schedule is a specification for a schedule resource
type Schedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *ScheduleSpec  `json:"spec"`
	Status ScheduleStatus `json:"status"`
}

// ScheduleSpec is the spec for a Schedule resource
type ScheduleSpec struct {
	Schedules []*ScheduleItem `json:"schedules"`
}

type ScheduleItem struct {
	Replicas int32      `json:"replicas"`
	Selector string     `json:"selector"`
	Start    *SchedSpan `json:"start"`
	Stop     *SchedSpan `json:"stop"`
}

type SchedSpan struct {
	Day  string    `json:"day"`
	Time *TimeSpan `json:"time"`
}

type TimeSpan struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// ScheduleStatus is the status for a Schedule resource
type ScheduleStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ScheduleList is a list of Schedule resources
type ScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Schedule `json:"items"`
}
