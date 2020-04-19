package traits

import (
	v1 "k8s.io/api/core/v1"
)

type ResourcesPolicy struct {
	Container string          `json:"container"`
	Limits    v1.ResourceList `json:"limits,omitempty" protobuf:"bytes,1,rep,name=limits,casttype=ResourceList,castkey=ResourceName"`
}
