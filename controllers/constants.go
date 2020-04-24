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
	Created         = "Created"
	Updated         = "Updated"
	Patched         = "Patched"
	Failed          = "Failed"
	Synced          = "Synced"
	SyncFailed      = "Sync Failed"
	SyncSuccessfuly = "Sync Successfully"
	Undefined       = "Undefined"
	NotFound        = "Not Found"

	// status
	PatchFailed  = "Patch Failed"
	CreateFailed = "Create Failed"
	Healthy      = "Healthy"
	Unhealthy    = "Unhealthy"
	// event messages
	MessageResourceExists  = "Resource %s/%s already exists and is not managed by Foo"
	MessageResourceCreated = "Resource %s/%s created successfully"
	MessageResourceUpdated = "Resource %s/%s updated successfully"
	MessageResourcePatched = "Resource %s/%s patched successfully"
	WorkeloadTypeUndefined = "Workload type %s is undefined"
	ComponentNotFound      = "ComponentSchematic %s not found"

	MessageResourceSynced = "ApplicationConfiguration synced successfully"

	//kind
	DeploymentKind   = "Deployment"
	ServiceKind      = "Service"
	IngressKind      = "Ingress"
	JobKind          = "Job"
	ConfigMapKind    = "ConfigMap"
	HpaKind          = "HorizontalPodAutoscaler"
	HcHpaKind        = "HorizontalPodAutoscaler"
	PvcKind          = "PersistentVolumeClaim"
	MysqlClusterKind = "MysqlCluster"

	ServerKind          = "Server"
	SingletonServerKind = "SingletonServer"
	WorkerKind          = "Worker"
	SingletonWorkerKind = "SingletonWorker"
	TaskKind            = "Task"
	SingletonTaskKind   = "SingletonTask"

	// group version
	OamV1alpha1GroupVersion  = "core.oam.dev/v1alpha1"
	MysqlClusterGroupVersion = "mysql.middleware.harmonycloud.cn/v1alpha1"

	// api version
	DeploymentApiVersion         = "apps/v1"
	ServiceApiVersion            = "v1"
	IngressApiVersion            = "extensions/v1beta1"
	JobApiVersion                = "batch/v1"
	ConfigMapApiVersion          = "v1"
	HpaApiVersion                = "autoscaling/v1"
	HcHpaApiVersion              = "harmonycloud.cn/v1beta1"
	PvcApiVersion                = "v1"
	MysqlClusterApiVersion       = "mysql.middleware.harmonycloud.cn/v1alpha1"
	ApplicationConfigurationKind = "ApplicationConfiguration"
	Component                    = "Component"

	// label
	Role     = "role"
	Instance = "instance"
	Workload = "workload"
	Trait    = "trait"

	// common
	Error = "Error"
)
