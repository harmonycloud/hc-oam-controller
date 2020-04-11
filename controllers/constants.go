package controllers

const (
	// workload types
	WorkloadTypeServer          = "core.oam.dev/v1alpha1.Server"
	WorkloadTypeSingletonServer = "core.oam.dev/v1alpha1.SingletonServer"
	WorkloadTypeWorker          = "core.oam.dev/v1alpha1.Worker"
	WorkloadTypeSingletonWorker = "core.oam.dev/v1alpha1.SingletonWorker"
	WorkloadTypeTask            = "core.oam.dev/v1alpha1.Task"
	WorkloadTypeSingletonTask   = "core.oam.dev/v1alpha1.SingletonTask"
	WorkloadTypeMysqlCluster    = "harmonycloud.cn/v1alpha1.MysqlCluster"

	// event reasons
	Created       = "Created"
	Updated       = "Updated"
	Patched       = "Patched"
	Failed        = "Failed"
	SuccessSynced = "Synced"

	// event messages
	MessageResourceExists  = "Resource %s/%s already exists and is not managed by Foo"
	MessageResourceCreated = "Resource %s/%s created successfully"
	MessageResourceUpdated = "Resource %s/%s updated successfully"
	MessageResourcePatched = "Resource %s/%s patched successfully"

	MessageResourceSynced = "ApplicationConfiguration synced successfully"
)
