package traits

type LogPilot struct {
	Container      string `json:"container"`
	Path           string `json:"path"`
	Name           string `json:"name"`
	Tags           string `json:"tags,omitempty"`
	PilotLogPrefix string `json:"pilotLogPrefix,omitempty"`
}
