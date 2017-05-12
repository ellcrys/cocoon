package scheduler

import (
	"fmt"
)

// Job defines a nomad job specification
type Job struct {
	Region      string
	ID          string
	Name        string
	Type        string
	Priority    int
	AllAtOnce   bool
	Datacenters []string
	Constraints []Constraint
	TaskGroups  []TaskGroup
	Update      Update
}

// Constraint defines a job/task contraint
type Constraint struct {
	LTarget string
	RTarget string
	Operand string
}

// Update defines the update stanza
type Update struct {
	Stagger     int64
	MaxParallel int
}

// TaskGroup defines the task_group stanza
type TaskGroup struct {
	Name          string
	Count         int
	Constraints   []string
	Tasks         []Task
	EphemeralDisk EphemeralDisk
	RestartPolicy RestartPolicy
	Meta          map[string]string
}

// Task defines a job task
type Task struct {
	Name            string
	Driver          string
	Config          Config
	Env             map[string]string
	Services        []NomadService
	Meta            map[string]string
	LogConfig       LogConfig
	KillTimeout     int64
	Templates       []Template
	Artifacts       []Artifact
	Resources       Resources
	DispatchPayload DispatchPayload
}

// DispatchPayload configures tast to have access to dispatch payload
type DispatchPayload struct {
	File string
}

// Resources defines the resources to allocate to a task
type Resources struct {
	CPU      int
	MemoryMB int
	IOPS     int
	DiskMB   int
	Networks []Network
}

// Network defines network allocation
type Network struct {
	MBits        int
	DynamicPorts []DynamicPort
}

// RestartPolicy defines restart policy
type RestartPolicy struct {
	Interval int64
	Attempts int
	Delay    int64
	Mode     string
}

// DynamicPort defines a dynamic port allocation
type DynamicPort struct {
	Label string
}

// Artifact defines an artifact to be downloaded
type Artifact struct {
	GetterSource  string
	RelativeDest  string
	GetterOptions map[string]string
}

// Template defines template objects to render for the task
type Template struct {
	SourcePath   string
	DestPath     string
	EmbeddedTmpl string
	ChangeMode   string
	ChangeSignal string
	Splay        int64
}

// LogConfig defines log configurations
type LogConfig struct {
	MaxFiles      int
	MaxFileSizeMB int
}

// NomadService defines a service
type NomadService struct {
	Name      string
	Tags      []string
	PortLabel string
	Checks    []Check
}

// Check defines a service check
type Check struct {
	ID       string `json:"Id"`
	Name     string
	Type     string
	Path     string
	Port     string
	Timeout  int64
	Interval int64
	Protocol string
}

// Logging defines logging configuration
type Logging struct {
	Type   string              `json:"type"`
	Config []map[string]string `json:"config"`
}

// Config defines a driver/task configuration
type Config struct {
	NetworkMode string    `json:"network_mode"`
	Privileged  bool      `json:"privileged"`
	ForcePull   bool      `json:"force_pull"`
	Volumes     []string  `json:"volumes"`
	Image       string    `json:"image"`
	Command     string    `json:"command"`
	Args        []string  `json:"args"`
	Logging     []Logging `json:"logging"`
}

// EphemeralDisk is an ephemeral disk object
type EphemeralDisk struct {
	Sticky  bool
	Migrate bool
	SizeMB  int `mapstructure:"size,omitempty"`
}

// NomadJob represents a nomad job
type NomadJob struct {
	Job *Job
}

// NewJob creates a new job with some default values.
func NewJob(version, id string, count int) *NomadJob {
	return &NomadJob{
		Job: &Job{
			Region:      "",
			ID:          id,
			Name:        id,
			Type:        "service",
			Priority:    50,
			AllAtOnce:   false,
			Datacenters: []string{},
			Constraints: []Constraint{
				Constraint{
					LTarget: "${attr.kernel.name}",
					RTarget: "linux",
					Operand: "=",
				},
			},
			TaskGroups: []TaskGroup{
				TaskGroup{
					Name:  fmt.Sprintf("cocoon-grp-%s", id),
					Count: count,
					Meta: map[string]string{
						"VERSION":   version,
						"REPO_USER": "ncodes",
					},
					Tasks: []Task{
						Task{
							Name:   "connector",
							Driver: "docker",
							Config: Config{
								Image:       "${NOMAD_META_REPO_USER}/cocoon-launcher:latest",
								NetworkMode: "host",
								Privileged:  true,
								ForcePull:   true,
								Volumes: []string{
									"/var/run/docker.sock:/var/run/docker.sock",
								},
								Command: "bash",
								Args:    []string{"/local/run-connector.sh"},
							},
							Env: map[string]string{
								"ROUTER_DOMAIN":         "whatbay.co",
								"VERSION":               "${NOMAD_META_VERSION}",
								"ENV":                   "",
								"COCOON_ID":             id,
								"COCOON_RELEASE":        "",
								"COCOON_DISK_LIMIT":     "",
								"COCOON_CONTAINER_NAME": "code-${NOMAD_ALLOC_ID}",
							},
							Services: []NomadService{
								NomadService{
									Name:      fmt.Sprintf("cocoons"),
									Tags:      []string{id},
									PortLabel: "RPC",
								},
							},
							Meta: map[string]string{},
							LogConfig: LogConfig{
								MaxFiles:      10,
								MaxFileSizeMB: 10,
							},
							Templates: []Template{},
							Artifacts: []Artifact{
								Artifact{
									GetterSource: "https://raw.githubusercontent.com/${NOMAD_META_REPO_USER}/cocoon/master/scripts/run-connector.sh",
									RelativeDest: "/local",
								},
							},
							KillTimeout: 15000000000,
							Resources: Resources{
								CPU:      100,
								MemoryMB: 1024,
								DiskMB:   800,
								IOPS:     0,
								Networks: []Network{
									Network{
										MBits: 1,
										DynamicPorts: []DynamicPort{
											DynamicPort{Label: "RPC"},
											DynamicPort{Label: "HTTP"},
										},
									},
								},
							},
							DispatchPayload: DispatchPayload{},
						},
						Task{
							Name:   "code",
							Driver: "docker",
							Config: Config{
								Image:       "${NOMAD_META_REPO_USER}/launch-go:latest",
								NetworkMode: "bridge",
								ForcePull:   true,
								Command:     "bash",
								Args:        []string{"-c", "echo 'Hello Human. I am alive'; tail -f /dev/null"},
							},
							Env:      map[string]string{},
							Services: []NomadService{},
							Meta:     map[string]string{},
							LogConfig: LogConfig{
								MaxFiles:      10,
								MaxFileSizeMB: 10,
							},
							Templates:   []Template{},
							Artifacts:   []Artifact{},
							KillTimeout: 15000000000,
							Resources: Resources{
								CPU:      100,
								MemoryMB: 1024,
								DiskMB:   800,
								IOPS:     0,
								Networks: []Network{
									Network{
										MBits: 1,
										DynamicPorts: []DynamicPort{
											DynamicPort{Label: "RPC"},
										},
									},
								},
							},
							DispatchPayload: DispatchPayload{},
						},
					},
					EphemeralDisk: EphemeralDisk{
						Sticky:  false,
						SizeMB:  0,
						Migrate: false,
					},
					RestartPolicy: RestartPolicy{
						Interval: 300000000000,
						Attempts: 10,
						Delay:    25000000000,
						Mode:     "delay",
					},
				},
			},
			Update: Update{
				Stagger:     10000000000,
				MaxParallel: 1,
			},
		},
	}
}

// GetSpec returns the job's specification
func (j *NomadJob) GetSpec() *Job {
	return j.Job
}

//SetVersion set the connectors version
func (j *NomadJob) SetVersion(v string) {
	j.GetSpec().TaskGroups[0].Meta["VERSION"] = v
}
