package traits

import "k8s.io/apimachinery/pkg/util/intstr"

type Ingress struct {
	Hostname     string             `json:"hostname"`
	ServicePort  intstr.IntOrString `json:"servicePort"`
	Path         string             `json:"path,omitempty"`
	IngressClass string             `json:"ingressClass,omitempty"`
}
