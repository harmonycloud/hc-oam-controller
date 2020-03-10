package traits

type AutoScaler struct {
	Minimum string `json:"minimum"`
	Maximum string `json:"maximum"`
	Memory  string `json:"memory,omitempty"`
	Cpu     string `json:"cpu,omitempty"`
}
