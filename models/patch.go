package loggerInjector

import v1 "k8s.io/api/core/v1"

type PatchPayload[T Spec | CronJobSpec] struct {
	Spec T `json:"spec"`
}
type Spec struct {
	Template Template `json:"template"`
}
type Template struct {
	Spec     *v1.PodSpec `json:"spec"`
	MetaData MetaData    `json:"metadata"`
}
type MetaData struct {
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}
type JobTemplate struct {
	Spec Spec `json:"spec"`
}

type CronJobSpec struct {
	JobTemplate JobTemplate `json:"jobTemplate"`
}
