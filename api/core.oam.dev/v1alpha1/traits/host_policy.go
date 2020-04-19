package traits

type HostPolicy struct {
	HostNetwork bool `json:"hostNetwork"`
	HostPid     bool `json:"hostPid"`
	HostIpc     bool `json:"hostIpc"`
}
