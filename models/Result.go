package loggerInjector

import v1 "k8s.io/api/core/v1"

type Result struct {
	Name        string
	Namespace   string
	Annotations map[string]string
	Labels      map[string]string
	Spec        *v1.PodSpec
	Conditions  []Condition
}

type Condition struct {
	Status v1.ConditionStatus
	Type   string
}
