package compose

import (
	"strconv"
	"strings"
)

type ShellCommand []string

// 服务配置
// https://github.com/compose-spec/compose-go/blob/b6c7aa633e9103114a99707c80184c5b8cbbb35f/types/types.go
// https://docs.docker.com/compose/compose-file/compose-file-v3/
type Service struct {
	Name     string   `yaml:"-" json:"-"`
	Profiles []string `mapstructure:"profiles" yaml:"profiles,omitempty" json:"profiles,omitempty"`

	Build Build `yaml:"build,omitempty"` // 构建时的配置项
	//BlkioConfig     *BlkioConfig                     `mapstructure:"blkio_config" yaml:",omitempty" json:"blkio_config,omitempty"`
	CapAdd       []string `yaml:"cap_add,omitempty"`       // 添加的容器功能
	CapDrop      []string `yaml:"cap_drop,omitempty"`      // 添加的容器功能
	CgroupParent string   `yaml:"cgroup_parent,omitempty"` // 可选的父控制组
	Command      []string `yaml:"command,omitempty"`       // 默认命令
	CPUCount     int64    `mapstructure:"cpu_count" yaml:"cpu_count,omitempty" json:"cpu_count,omitempty"`
	CPUPercent   float32  `mapstructure:"cpu_percent" yaml:"cpu_percent,omitempty" json:"cpu_percent,omitempty"`
	CPUPeriod    int64    `mapstructure:"cpu_period" yaml:"cpu_period,omitempty" json:"cpu_period,omitempty"`
	CPUQuota     int64    `mapstructure:"cpu_quota" yaml:"cpu_quota,omitempty" json:"cpu_quota,omitempty"`
	CPURTPeriod  int64    `mapstructure:"cpu_rt_period" yaml:"cpu_rt_period,omitempty" json:"cpu_rt_period,omitempty"`
	CPURTRuntime int64    `mapstructure:"cpu_rt_runtime" yaml:"cpu_rt_runtime,omitempty" json:"cpu_rt_runtime,omitempty"`
	CPUS         float32  `mapstructure:"cpus" yaml:"cpus,omitempty" json:"cpus,omitempty"`
	CPUSet       string   `mapstructure:"cpuset" yaml:"cpuset,omitempty" json:"cpuset,omitempty"`
	CPUShares    int64    `mapstructure:"cpu_shares" yaml:"cpu_shares,omitempty" json:"cpu_shares,omitempty"`
	//Configs        []ServiceConfigObjConfig `yaml:",omitempty" json:"configs,omitempty"`
	//CredentialSpec *CredentialSpecConfig    `mapstructure:"credential_spec" yaml:"credential_spec,omitempty" json:"credential_spec,omitempty"`
	ContainerName string      `yaml:"container_name,omitempty"` // 容器名称
	DependsOn     interface{} `yaml:"depends_on,omitempty"`     // 服务之间的依赖关系 //TODO 有时候是[]string, 有时候是 map[string]interface{}
	DNS           []string    `yaml:"dns,omitempty"`
	//Deploy         *DeployConfig            `yaml:",omitempty" json:"deploy,omitempty"`
	Devices    []string `yaml:",omitempty" json:"devices,omitempty"`
	DNSOpts    []string `mapstructure:"dns_opt" yaml:"dns_opt,omitempty" json:"dns_opt,omitempty"`
	DNSSearch  []string `mapstructure:"dns_search" yaml:"dns_search,omitempty" json:"dns_search,omitempty"`
	Dockerfile string   `yaml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	DomainName string   `mapstructure:"domainname" yaml:"domainname,omitempty" json:"domainname,omitempty"`

	Environment map[string]string `yaml:"environment,omitempty"` // 环境变量
	Expose      []string          `yaml:"expose,omitempty"`      // 暴露的端口
	Entrypoint  ShellCommand      `yaml:",omitempty" json:"entrypoint,omitempty"`
	EnvFile     []string          `mapstructure:"env_file" yaml:"env_file,omitempty" json:"env_file,omitempty"`
	//Extends        ExtendsConfig            `yaml:"extends,omitempty" json:"extends,omitempty"`
	ExternalLinks []string `mapstructure:"external_links" yaml:"external_links,omitempty" json:"external_links,omitempty"`
	ExtraHosts    []string `mapstructure:"extra_hosts" yaml:"extra_hosts,omitempty" json:"extra_hosts,omitempty"`

	GroupAdd    []string               `mapstructure:"group_add" yaml:"group_add,omitempty" json:"group_add,omitempty"`
	Hostname    string                 `yaml:",omitempty" json:"hostname,omitempty"`
	HealthCheck map[string]interface{} `yaml:"healthcheck,omitempty"` // 健康检查
	Image       string                 `yaml:"image"`                 // 容器启动的镜像
	Init        *bool                  `yaml:",omitempty" json:"init,omitempty"`
	Ipc         string                 `yaml:",omitempty" json:"ipc,omitempty"`
	Isolation   string                 `mapstructure:"isolation" yaml:"isolation,omitempty" json:"isolation,omitempty"`
	Labels      map[string]string      `yaml:",omitempty" json:"labels,omitempty"`
	Links       []string               `yaml:",omitempty" json:"links,omitempty"`
	//Logging         *LoggingConfig                   `yaml:",omitempty" json:"logging,omitempty"`
	LogDriver string            `mapstructure:"log_driver" yaml:"log_driver,omitempty" json:"log_driver,omitempty"`
	LogOpt    map[string]string `mapstructure:"log_opt" yaml:"log_opt,omitempty" json:"log_opt,omitempty"`
	//MemLimit        UnitBytes                        `mapstructure:"mem_limit" yaml:"mem_limit,omitempty" json:"mem_limit,omitempty"`
	//MemReservation  UnitBytes                        `mapstructure:"mem_reservation" yaml:"mem_reservation,omitempty" json:"mem_reservation,omitempty"`
	//MemSwapLimit    UnitBytes                        `mapstructure:"memswap_limit" yaml:"memswap_limit,omitempty" json:"memswap_limit,omitempty"`
	//MemSwappiness   UnitBytes                        `mapstructure:"mem_swappiness" yaml:"mem_swappiness,omitempty" json:"mem_swappiness,omitempty"`
	MacAddress  string   `mapstructure:"mac_address" yaml:"mac_address,omitempty" json:"mac_address,omitempty"`
	Net         string   `yaml:"net,omitempty" json:"net,omitempty"`
	NetworkMode string   `mapstructure:"network_mode" yaml:"network_mode,omitempty" json:"network_mode,omitempty"`
	Networks    []string `yaml:"networks,omitempty"` // 加入的网络

	OomKillDisable bool     `mapstructure:"oom_kill_disable" yaml:"oom_kill_disable,omitempty" json:"oom_kill_disable,omitempty"`
	OomScoreAdj    int64    `mapstructure:"oom_score_adj" yaml:"oom_score_adj,omitempty" json:"oom_score_adj,omitempty"`
	Pid            string   `yaml:",omitempty" json:"pid,omitempty"`
	PidsLimit      int64    `mapstructure:"pids_limit" yaml:"pids_limit,omitempty" json:"pids_limit,omitempty"`
	Platform       string   `yaml:",omitempty" json:"platform,omitempty"`
	Ports          []string `yaml:"ports,omitempty"` // 暴露的端口号

	Restart string   `yaml:"restart,omitempty"` // 重启策略
	Secrets []Secret `yaml:"secrets,omitempty"` // 密钥

	Privileged      bool              `yaml:",omitempty" json:"privileged,omitempty"`
	PullPolicy      string            `mapstructure:"pull_policy" yaml:"pull_policy,omitempty" json:"pull_policy,omitempty"`
	ReadOnly        bool              `mapstructure:"read_only" yaml:"read_only,omitempty" json:"read_only,omitempty"`
	Runtime         string            `yaml:",omitempty" json:"runtime,omitempty"`
	Scale           int               `yaml:"-" json:"-"`
	SecurityOpt     []string          `mapstructure:"security_opt" yaml:"security_opt,omitempty" json:"security_opt,omitempty"`
	ShmSize         int64             `mapstructure:"shm_size" yaml:"shm_size,omitempty" json:"shm_size,omitempty"`
	StdinOpen       bool              `mapstructure:"stdin_open" yaml:"stdin_open,omitempty" json:"stdin_open,omitempty"`
	StopGracePeriod *int64            `mapstructure:"stop_grace_period" yaml:"stop_grace_period,omitempty" json:"stop_grace_period,omitempty"`
	StopSignal      string            `mapstructure:"stop_signal" yaml:"stop_signal,omitempty" json:"stop_signal,omitempty"`
	Sysctls         map[string]string `yaml:",omitempty" json:"sysctls,omitempty"`
	Tmpfs           int64             `yaml:",omitempty" json:"tmpfs,omitempty"`
	Tty             bool              `mapstructure:"tty" yaml:"tty,omitempty" json:"tty,omitempty"`
	//Ulimits         map[string]*UlimitsConfig `yaml:",omitempty" json:"ulimits,omitempty"`

	User         string   `yaml:",omitempty" json:"user,omitempty"` // 执行的用户
	UserNSMode   string   `mapstructure:"userns_mode" yaml:"userns_mode,omitempty" json:"userns_mode,omitempty"`
	Uts          string   `yaml:"uts,omitempty" json:"uts,omitempty"`
	VolumeDriver string   `mapstructure:"volume_driver" yaml:"volume_driver,omitempty" json:"volume_driver,omitempty"`
	VolumesFrom  []string `mapstructure:"volumes_from" yaml:"volumes_from,omitempty" json:"volumes_from,omitempty"`
	WorkingDir   string   `mapstructure:"working_dir" yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
	Volumes      []string `yaml:"volumes,omitempty"` // 挂载卷
}

func (service *Service) AppendPort(port string) {
	service.Ports = append(service.Ports, port)
}

// 是否包含某个容器端口的映射
func (service *Service) ContainContainerPort(port int) bool {
	originPort := strconv.Itoa(port)
	for _, portStr := range service.Ports {
		array := strings.Split(portStr, ":")
		if array[1] == originPort {
			return true
		}
	}

	return false
}
