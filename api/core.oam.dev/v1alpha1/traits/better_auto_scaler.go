package traits

type BetterAutoScaler struct {
	Minimum    int32 `json:"minimum"`
	Maximum    int32 `json:"maximum"`
	MemoryUp   int32 `json:"memory-up,omitempty"`
	MemoryDown int32 `json:"memory-down,omitempty"`
	CpuUp      int32 `json:"cpu-up,omitempty"`
	CpuDown    int32 `json:"cpu-down,omitempty"`
}
