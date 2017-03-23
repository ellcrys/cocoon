package scheduler

import (
	"fmt"
	"strconv"

	"os"

	"github.com/ellcrys/crypto"
	"github.com/ellcrys/util"
	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/common"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("nomad")

// SupportedCocoonCodeLang defines the supported chaincode language
var SupportedCocoonCodeLang = []string{"go"}

// SupportedMemory represents the allowed cocoon memory choices
var SupportedMemory = map[string]int{
	"512m": 512,
	"1g":   1024,
	"2g":   2048,
}

// SupportedCPUShare represents the allowed cocoon cpu share choices
var SupportedCPUShare = map[string]int{
	"1x": 100,
	"2x": 200,
}

// SupportedDiskSpace represents the allowed cocoon disk space
var SupportedDiskSpace = map[string]int{
	"1x": 300,
	"2x": 500,
}

// Nomad defines a nomad scheduler that implements
// scheduler.Scheduler interface. Every interaction with
// the scheduler is handled here.
type Nomad struct {
	schedulerAddr    string
	API              string
	ServiceDiscovery ServiceDiscovery
}

// NewNomad creates a nomad scheduler object
func NewNomad() *Nomad {
	return &Nomad{
		ServiceDiscovery: &NomadServiceDiscovery{
			ConsulAddr: util.Env("CONSUL_ADDR", "127.0.0.7:4646"),
			Protocol:   "http",
		},
	}
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

	respStr, _ := res.Body.ToString()
	return respStr, res.StatusCode, nil
}

// Deploy a cocoon code to the scheduler
func (sc *Nomad) Deploy(jobID, lang, url, tag, buildParams, link, memory, cpuShare string) (*DeploymentInfo, error) {

	var err error

	if len(jobID) == 0 {
		return nil, fmt.Errorf("job id is required")
	}

	if err = common.ValidateDeployment(url, lang, buildParams); err != nil {
		return nil, err
	}

	log.Debugf("Deploying cocoon code with language=%s, url=%s, tag=%s", lang, url, tag)

	if len(buildParams) > 0 {
		buildParams = crypto.ToBase64([]byte(buildParams))
	}

	var img string
	switch lang {
	case "go":
		img = "ncodes/cocoon-launcher:latest"
	}

	job := NewJob(jobID, 1)
	job.GetSpec().Region = "global"
	job.GetSpec().Datacenters = []string{"dc1"}
	job.GetSpec().TaskGroups[0].Tasks[0].Env["COCOON_CODE_URL"] = url
	job.GetSpec().TaskGroups[0].Tasks[0].Env["COCOON_CODE_TAG"] = tag
	job.GetSpec().TaskGroups[0].Tasks[0].Env["COCOON_CODE_LANG"] = lang
	job.GetSpec().TaskGroups[0].Tasks[0].Env["COCOON_BUILD_PARAMS"] = buildParams
	job.GetSpec().TaskGroups[0].Tasks[0].Env["COCOON_DISK_LIMIT"] = strconv.Itoa(SupportedDiskSpace[cpuShare])
	job.GetSpec().TaskGroups[0].Tasks[0].Env["COCOON_LINK"] = link
	job.GetSpec().TaskGroups[0].Tasks[0].Config.Image = img
	job.GetSpec().TaskGroups[0].Tasks[0].Resources.CPU = SupportedCPUShare[cpuShare]
	job.GetSpec().TaskGroups[0].Tasks[0].Resources.MemoryMB = SupportedMemory[memory]
	job.GetSpec().TaskGroups[0].Resources.CPU = SupportedCPUShare[cpuShare]
	job.GetSpec().TaskGroups[0].Resources.MemoryMB = SupportedMemory[memory]

	jobSpec, _ := util.ToJSON(job)
	resp, status, err := sc.deployJob(string(jobSpec))
	if err != nil {
		return nil, fmt.Errorf("system: failed to deploy job spec. %s", err)
	} else if status != 200 {
		return nil, fmt.Errorf("system: failed to deploy job spec. %s", resp)
	}

	var jobInfo map[string]interface{}
	if err = util.FromJSON([]byte(resp), &jobInfo); err != nil {
		return nil, fmt.Errorf("system: %s", resp)
	}

	return &DeploymentInfo{
		ID:     jobID,
		EvalID: jobInfo["EvalID"].(string),
	}, nil
}

// Getenv returns an environment variable value based on the schedulers
// naming convention.
func Getenv(env string) string {
	return os.Getenv("NOMAD_" + env)
}

// GetServices fetches all the instances of a service
func (sc *Nomad) GetServices(serviceID string) []Service {
	return nil
}
