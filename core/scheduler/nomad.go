package scheduler

import (
	"fmt"

	"os"

	"github.com/ellcrys/cocoon/core/common"
	"github.com/ellcrys/cocoon/core/config"
	"github.com/ellcrys/util"
	"github.com/franela/goreq"
	"github.com/hashicorp/consul/api"
)

var log = config.MakeLogger("nomad")

// SupportedCocoonCodeLang defines the supported chaincode language
var SupportedCocoonCodeLang = []string{"go"}

// Nomad defines a nomad scheduler that implements
// scheduler.Scheduler interface. Every interaction with
// the scheduler is handled here.
type Nomad struct {
	schedulerAddr string
	API           string
}

// NewNomad creates a nomad scheduler object
func NewNomad() *Nomad {
	return &Nomad{}
}

// GetName returns the scheduler name
func (sc *Nomad) GetName() string {
	return "nomad"
}

// SetAddr sets the nomad's API endpoint
func (sc *Nomad) SetAddr(addr string, https bool) {
	scheme := "http://"
	if https {
		scheme = "https://"
	}
	sc.API = scheme + addr
}

// deployJob registers a new job
func (sc *Nomad) deployJob(jobSpec string) (string, int, error) {
	res, err := goreq.Request{
		Method: "POST",
		Uri:    sc.API + "/v1/jobs",
		Body:   jobSpec,
	}.Do()
	if err != nil {
		return "", 0, err
	}
	defer res.Body.Close()
	respStr, _ := res.Body.ToString()
	return respStr, res.StatusCode, nil
}

// makeLinkToServiceTag creates a tag representing a link to a cocoon id.
// To be used as a service tag
func makeLinkToServiceTag(linkID string) string {
	return fmt.Sprintf("link_to:%s", linkID)
}

// Deploy a cocoon code to the scheduler
func (sc *Nomad) Deploy(jobID, releaseID string, memory, cpuShare int) (*DeploymentInfo, error) {

	var err error

	if len(jobID) == 0 {
		return nil, fmt.Errorf("job id is required")
	}

	log.Debugf("Deploying cocoon=%s, release=%s", jobID, releaseID)

	// define job specification
	job := NewJob("master", jobID, 1)
	job.GetSpec().Region = "global"
	job.GetSpec().Datacenters = []string{"dc1"}
	job.GetSpec().TaskGroups[0].Tasks[0].Env["ENV"] = os.Getenv("ENV")
	job.GetSpec().TaskGroups[0].Tasks[0].Env["COCOON_RELEASE"] = releaseID
	job.GetSpec().TaskGroups[0].Tasks[0].Resources.CPU = cpuShare
	job.GetSpec().TaskGroups[0].Tasks[0].Resources.DiskMB = 1000
	job.GetSpec().TaskGroups[0].Tasks[0].Resources.MemoryMB = common.Round(0.3 * float64(memory))
	job.GetSpec().TaskGroups[0].Tasks[1].Resources.MemoryMB = common.Round(0.7 * float64(memory))
	job.GetSpec().TaskGroups[0].EphemeralDisk.SizeMB = 4000
	job.GetSpec().TaskGroups[0].Tasks[0].Env["COCOON_DISK_LIMIT"] = "4000"

	// set fluentd logger in production environment
	if os.Getenv("ENV") == "production" {
		job.SetVersion(os.Getenv("CONNECTOR_VERSION"))
		job.GetSpec().TaskGroups[0].Tasks[0].Config.Logging = []Logging{{
			Type: "fluentd",
			Config: []map[string]string{{
				"fluentd-address": "localhost:24224",
				"tag":             fmt.Sprintf("cocoon-%s", jobID),
			}}},
		}
	}

	// set shared volume
	job.AssignSharedVolume()
	log.Debug("Shared directory assigned")

	// deploy job specification
	jobSpec, _ := util.ToJSON(job)
	resp, status, err := sc.deployJob(string(jobSpec))
	if err != nil {
		return nil, fmt.Errorf("system: failed to deploy job spec. %s", err)
	} else if status != 200 {
		return nil, fmt.Errorf("system: failed to deploy job spec. %s", resp)
	}

	return &DeploymentInfo{
		ID: jobID,
	}, nil
}

// GetDeploymentStatus gets the status of a job
func (sc *Nomad) GetDeploymentStatus(jobID string) (string, error) {
	res, err := goreq.Request{
		Method: "GET",
		Uri:    sc.API + "/v1/job/" + jobID,
	}.Do()

	if err != nil {
		return "", err
	} else if res.StatusCode != 200 {
		respStr, _ := res.Body.ToString()
		res.Body.Close()
		if res.StatusCode == 404 {
			return "", fmt.Errorf("not found")
		}
		return "", fmt.Errorf(respStr)
	}

	defer res.Body.Close()
	var job map[string]interface{}
	err = res.Body.FromJsonTo(&job)
	if err != nil {
		return "", common.JSONCoerceErr("job", err)
	}

	if status, ok := job["Status"].(string); ok {
		return status, nil
	}

	return "", nil
}

// Stop stops a running cocoon job
func (sc *Nomad) Stop(jobID string) error {
	res, err := goreq.Request{
		Method: "DELETE",
		Uri:    sc.API + "/v1/job/" + jobID,
	}.Do()
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		respStr, _ := res.Body.ToString()
		res.Body.Close()
		return fmt.Errorf(respStr)
	}
	res.Body.Close()
	log.Infof("Successfully stopped cocoon = %s", jobID)
	return nil
}

// Getenv returns an environment variable value based on the schedulers
// naming convention.
func Getenv(env, defaultVal string) string {
	return util.Env("NOMAD_"+env, defaultVal)
}

// GetServiceDiscoverer returns an instance of the nomad service discovery
func (sc *Nomad) GetServiceDiscoverer() (ServiceDiscovery, error) {
	cfg := api.DefaultConfig()
	cfg.Address = util.Env("CONSUL_ADDR", cfg.Address)
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}
	_, err = client.Status().Leader()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}
	return NewNomadServiceDiscovery(client), nil
}
