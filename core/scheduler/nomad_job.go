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

// TaskGroups defines the task_group stanza
type TaskGroup struct {
	Name          string
	Count         int
	Constraints   []string
	Tasks         []Task
	Resources     Resources
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

// Config defines a driver/task configuration
type Config struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// NomadJob represents a nomad job
type NomadJob struct {
	Job *Job
}

// NewJob creates a new job with some default values.
func NewJob(connectorVersion, id string, count int) *NomadJob {
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
					Name:  fmt.Sprintf("tskgrp-%s", id),
					Count: count,
					Tasks: []Task{
						Task{
							Name:   fmt.Sprintf("task-%s", id),
							Driver: "raw_exec",
							Config: Config{
								Command: "bash",
								Args:    []string{"runner.sh"},
							},
							Env: map[string]string{
								"CONNECTOR_VERSION":   connectorVersion,
								"COCOON_ID":           id,
								"CONTAINER_ID":        id,
								"COCOON_CODE_URL":     "",
								"COCOON_CODE_TAG":     "",
								"COCOON_CODE_LANG":    "",
								"COCOON_BUILD_PARAMS": "",
								"COCOON_DISK_LIMIT":   "",
								// The name of the connector runner script and a link to the script.
								// The runner script will fetch and run whatever is found in this environment vars.
								"RUN_SCRIPT_NAME": "run-connector.sh",
								"RUN_SCRIPT_URL":  "https://rawgit.com/ncodes/cocoon/master/scripts/run-connector.sh",
							},
							Services: []NomadService{
								NomadService{
									Name:      fmt.Sprintf("cocoons-%s", id),
									Tags:      []string{id},
									PortLabel: "CONNECTOR_RPC",
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
									GetterSource: "https://rawgit.com/ncodes/cocoon/master/scripts/runner.sh",
								},
							},
							Resources: Resources{
								CPU:      20,
								MemoryMB: 10,
								IOPS:     0,
								Networks: []Network{
									Network{
										MBits: 1,
										DynamicPorts: []DynamicPort{
											DynamicPort{Label: "CONNECTOR_RPC"},
											DynamicPort{Label: "COCOON_RPC"},
										},
									},
								},
							},
							DispatchPayload: DispatchPayload{},
						},
					},
					Resources: Resources{
						CPU:      20,
						MemoryMB: 10,
						IOPS:     0,
						Networks: []Network{},
					},
					RestartPolicy: RestartPolicy{
						Interval: 300000000000,
						Attempts: 10,
						Delay:    25000000000,
						Mode:     "delay",
					},
					Meta: map[string]string{},
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
