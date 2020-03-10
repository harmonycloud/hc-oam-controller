package traits

type Ingress struct {
	Hostname    string `json:"hostname"`
	ServicePort string `json:"servicePort"`
	Path        string `json:"path,omitempty"`
}
