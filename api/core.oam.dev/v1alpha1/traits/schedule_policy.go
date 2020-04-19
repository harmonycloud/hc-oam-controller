package traits

type SchedulePolicy struct {
	NodeAffinity    Affinity `json:"nodeAffinity"`
	PodAffinity     Affinity `json:"podAffinity"`
	PodAntiAffinity Affinity `json:"podAntiAffinity"`
}

type Affinity struct {
	// value: required, preferred
	Type     string            `json:"type"`
	Selector map[string]string `json:"selector"`
}
