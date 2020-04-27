package traits

import "k8s.io/apimachinery/pkg/util/intstr"

type BetterAutoScaler struct {
	Minimum    intstr.IntOrString `json:"minimum"`
	Maximum    intstr.IntOrString `json:"maximum"`
	MemoryUp   intstr.IntOrString `json:"memory-up,omitempty"`
	MemoryDown intstr.IntOrString `json:"memory-down,omitempty"`
	CpuUp      intstr.IntOrString `json:"cpu-up,omitempty"`
	CpuDown    intstr.IntOrString `json:"cpu-down,omitempty"`
}
