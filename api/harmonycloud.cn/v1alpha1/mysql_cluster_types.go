package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 集群类型
type MysqlClusterType string

// 集群部署类型
type MysqlClusterDeployStrategyType string

// 集群当前阶段
type MysqlClusterPhase string

// 集群节点类型
type MysqlType string

// 集群节点读写模式
type MysqlMode string

const (
	MasterMasterReplication MysqlClusterType = "master-master"
	MasterSlaveReplication                   = "master-slave"

	AutoDeploy    MysqlClusterDeployStrategyType = "AutoDeploy"
	MigrateDeploy MysqlClusterDeployStrategyType = "MigrateDeploy"

	ClusterPhaseNone     MysqlClusterPhase = ""
	ClusterPhaseCreating                   = "Creating"
	ClusterPhaseRunning                    = "Running"
	ClusterPhaseFailed                     = "Failed"
	ClusterPhaseError                      = "Error"

	MysqlClusterName string    = "MysqlCluster"
	MasterMysql      MysqlType = "Master"
	SlaveMysql       MysqlType = "Slave"
	AllMysql         MysqlType = "All"
	UnknownMysql     MysqlType = "Unknown"

	ReadAndWrite MysqlMode = "R/W"
	ReadAndOnly  MysqlMode = "R/O"
	PodUnknown   MysqlMode = "Unknown"

	MysqlMiddleWareNameKey         = "mysql.middleware.harmonycloud.cn"
	PauseMiddleWareNameKey         = "pause.harmonycloud.cn"
	ReadOnlyMysqlMiddleWareNameKey = "readonly.mysql.middleware.harmonycloud.cn"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MysqlCluster struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec MysqlClusterSpec `json:"spec,omitempty"`

	Status MysqlClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MysqlClusterList struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []MysqlCluster `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// mysql集群部署规范
type MysqlClusterSpec struct {

	// label筛选
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// 集群期望实例
	Replicas *int32 `json:"replicas,omitempty"`

	// 集群是否进行维修状态，true代表进入，false代表不进入，默认应为false
	Paused bool `json:"paused,omitempty"`

	// 镜像地址，不同集群会配置不同地址
	Repository string `json:"repository,omitempty"`

	//切换主从
	SwitchMaster ClusterSwitch `json:"clusterSwitch, omitempty"`

	// 镜像版本，根据镜像版本进行特化处理
	Version string `json:"version,omitempty"`

	// 集群类型，分为主主同步、主从同步
	Type MysqlClusterType `json:"type,omitempty"`

	// 集群部署方式
	DeployStrategy MysqlClusterDeployStrategy `json:"deployStrategy,omitempty"`

	// 业务部署
	BusinessDeploy []BusinessDeploy `json:"businessDeploy, omitempty"`

	// statefulset 模板
	Statefulset StatefulSetPolicy `json:"statefulset,omitempty"`

	//迁移回迁信息
	MigratePolicy MigratePolicy `json:"migratePolicy,omitempty"`

	//强制主从切换
	ForceSwitched bool `json:"forceSwitched,omitempty"`
	//限制强制主从切换次数
	SwitchedNum int32 `json:"switchedNum,omitempty"`
	//是否开启被动限制
	PassiveSwitched bool `json:"passiveSwitched,omitempty"`
	//被动限制后是否立即生效
	PassiveAvail bool `json:"passiveAvail,omitempty"`
	//pvcName 业务端创建的pvc名称
	PvcName string `json:"pvcName,omitempty"`
	//volumeQuota 记录pvc容量
	VolumeQuota string `json:"volumeQuota,omitempty"`
	//CMName
	CmName string `json:"cmName,omitempty"`
	//Secret
	SecretName string `json:"secretName,omitempty"`
	//delete
	Delete bool `json:"delete,omitempty"`
	//Secret
	ProblemPodName string `json:"problemPodName,omitempty"`
	//Secret
	TargetNodeName string `json:"targetNodeName,omitempty"`

	//批量删除
	BatchDelete bool `json:"targetNodeName,omitempty"`
}

type MigratePolicy struct {
	Started      bool   `json:"started,omitempty"`
	Backed       bool   `json:"backed,omitempty"`
	BackFinished bool   `json:"backFinished,omitempty"`
	ReadWriteVip string `json:"readWriteVip,omitempty"`
	ReadOnlyVip  string `json:"readOnlyVip,omitempty"`
}

type BusinessDeploy struct {
	Database string `json:"database,omitempty"`
	User     string `json:"user,omitempty"`
	Pwd      string `json:"pwd,omitempty"`
}

type ClusterSwitch struct {

	// 主从切换操作是否完成
	Switched bool `json:"switched"`
	// 主动切换流程完结
	Finished bool `json:"finished,omitempty"`
	// 主动切换是否是强制的
	ForceSwitched bool `json:"forceSwitched,omitempty"`
	// 要进行切换的实例名
	Master    string `json:"master,omitempty"`
	SwitchPod string `json:"switchPod,omitempty"`
}

type StatefulSetPolicy struct {

	// 标签管理
	// 当前预估的标签： 租户标签、项目标签、集群应用标签
	Labels map[string]string `json:"labels,omitempty"`

	// 备注管理
	// 当前预估的备注：readonly.mysql.middleware.harmonycloud.cn(是否是只读域名)
	// mysql.middleware.harmonycloud.cn
	Annotations map[string]string `json:"annotations,omitempty"`

	// 亲和性调度
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	//服务暴露端口，默认：20001
	ServerPort int32 `json:"serverPort,omitempty"`

	//健康检查
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`

	// 资源配额
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// 环境变量配置
	// 当前预估的环境变量为ClusterName
	Env []corev1.EnvVar `json:"env,omitempty"`

	// 存储配置
	PersistentVolumeClaimSpec *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaimSpec,omitempty"`

	// 更新策略
	UpdateStrategy appsv1.StatefulSetUpdateStrategy `json:"updateStrategy,omitempty"`

	// 配置文件管理
	Configmap string `json:"configmap,omitempty"`

	// initcontainer镜像名称
	InitImage string `json:"initImage,omitempty"`

	// 监控container镜像名称
	MonitorImage string `json:"monitorImage,omitempty"`

	//主机密码
	Secret string `json:"secret,omitempty"`
}

type MysqlClusterDeployStrategy struct {

	// 集群部署类型
	Type MysqlClusterDeployStrategyType `json:"type,omitempty"`

	// 集群部署基础配置
	BasicConfig MysqlClusterBasicConfig `json:"basicConfig,omitempty"`

	// 集群迁移配置
	Migration MysqlClusterMigration `json:"migration,omitempty"`
}

// 集群部署基本配置
type MysqlClusterBasicConfig struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
}

// 集群迁移配置
type MysqlClusterMigration struct {
	TargetAddress  string `json:"targetAddress,omitempty"`
	TargetPort     int32  `json:"targetPort,omitempty"`
	TargetUser     string `json:"targetUser,omitempty"`
	TargetPassword string `json:"targetPassword,omitempty"`

	// 集群进行正式转移
	MigtatingStatus bool `json:"migtatingStatus,omitempty"`
}

type MysqlClusterStatus struct {
	// 当前集群实例数
	Replicas *int32 `json:"replicas,omitempty"`

	FailedCount int `json:"failedCount,omitempty"`

	// 集群当前状态
	Phase MysqlClusterPhase `json:"phase,omitempty"`

	Reason string `json:"reason,omitempty"`

	CurrentRevision string `json:"currentRevision,omitempty" protobuf:"bytes,6,opt,name=currentRevision"`

	UpdateRevision string `json:"updateRevision,omitempty" protobuf:"bytes,7,opt,name=updateRevision"`

	Conditions []MysqlClusterCondition `json:"conditions,omitempty"`

	CurrentSwitchedNum int32 `json:"currentSwitchedNum,omitempty"`
}

type MysqlClusterCondition struct {
	Name               string      `json:"name,omitempty"`
	Type               MysqlType   `json:"type,omitempty"`
	Mode               MysqlMode   `json:"mode,omitempty"`
	Status             bool        `json:"status,omitempty"`
	Reason             string      `json:"reason,omitempty"`
	Message            string      `json:"message,omitempty"`
	PodIP              string      `json:"podIP,omitempty"`
	NodeName           string      `json:"nodeName,omitempty"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

type LeaderElectionConfiguration struct {
	LeaderElect bool

	LeaseDuration metav1.Duration

	RenewDeadline metav1.Duration

	RetryPeriod metav1.Duration

	ResourceLock string

	LockObjectNamespace string

	LockObjectName string
}

type TransferMcCluster struct {
	//是否开始迁移
	Start string `json:"start,omitempty"`
	//迁移到某步骤
	Step string `json:"start,omitempty"`
	//是否失败
	Failed bool `json:"failed,omitempty"`
	//旧集群名称
	OldMcName string `json:"oldMcName,omitempty"`
	//当前集群名称
	CurrentMcName string `json:"currentMcName,omitempty"`
	//是否回切
	IsBack bool `json:"isBack,omitempty"`
}

type OperatorManagerConfig struct {
	metav1.TypeMeta

	// Operators is the list of operators to enable or disable
	// '*' means "all enabled by default operators"
	// 'foo' means "enable 'foo'"
	// '-foo' means "disable 'foo'"
	// first item for a particular name wins
	Operators []string

	// ConcurrentRedisClusterSyncs is the number of redisCluster objects that are
	// allowed to sync concurrently. Larger number = more responsive redisClusters,
	// but more CPU (and network) load.
	ConcurrentRedisClusterSyncs int32

	//cluster create or upgrade timeout (min)
	ClusterTimeOut int32

	// The number of old history to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Defaults to 2.
	// +optional
	RevisionHistoryLimit int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,6,opt,name=revisionHistoryLimit"`

	// How long to wait between starting controller managers
	ControllerStartInterval metav1.Duration

	ResyncPeriod int64
	// leaderElection defines the configuration of leader election client.
	LeaderElection LeaderElectionConfiguration
	// port is the port that the controller-manager's http service runs on.
	Port int32
	// address is the IP address to serve on (set to 0.0.0.0 for all interfaces).
	Address string

	// enableProfiling enables profiling via web interface host:port/debug/pprof/
	EnableProfiling bool
	// contentType is contentType of requests sent to apiserver.
	ContentType string
	// kubeAPIQPS is the QPS to use while talking with kubernetes apiserver.
	KubeAPIQPS float32
	// kubeAPIBurst is the burst to use while talking with kubernetes apiserver.
	KubeAPIBurst int32
	//WorkQueue for operator
	WorkQueue int32
	//operatorpwd
	OperatorPwd string
	//replicaPwd
	ReplicaPwd string
	//sshPwd
	SshPwd string
	//operatorLyZone-1
	OperatorLyZone string
	//operatorHaZone
	OperatorHaZone string
	//operatorName
	OperatorName string
	//etcd
	EtcdServer   string
	EtcdKeyFile  string
	EtcdCertFile string
	EtcdCaFile   string
}
