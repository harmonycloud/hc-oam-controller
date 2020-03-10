package traits

type VolumeMounter struct {
	VolumeName   string `json:"volumeName"`
	StorageClass string `json:"storageClass"`
}
