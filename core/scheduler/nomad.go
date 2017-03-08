package scheduler

import (
	"fmt"

	"github.com/ellcrys/crypto"
	"github.com/ellcrys/util"
	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/data"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("nomad")

// SupportedCocoonCodeLang defines the supported chaincode language
var SupportedCocoonCodeLang = []string{"go"}

// SupportedMemory represents the allowed cocoon memory choices
var SupportedMemory = map[string]interface{}{
	"512m": 512,
	"1g":   1024,
	"2g":   2048,
}

// SupportedCPUShare represents the allowed cocoon cpu share choices
var SupportedCPUShare = map[string]interface{}{
	"1x": 100,
	"2x": 200,
}

// SupportedDiskSpace represents the allowed cocoon disk space
var SupportedDiskSpace = map[string]interface{}{
	"1x": 300,
	"2x": 500,
}

// Nomad defines a nomad scheduler that implements
// scheduler.Scheduler interface. Every interaction with
// the scheduler is handled here.
type Nomad struct {
	schedulerAddr string
	API           string
}

// NewNomad creates a nomad scheduler object
func NewNomad() *Nomad {
	return new(Nomad)
}

// SetAddr sets the nomad's API endpoint
func (sc *Nomad) SetAddr(addr string, https bool) {
	scheme := "http://"
	if https {
		scheme = "https://"
	}
	sc.API = scheme + addr
}

// PrepareJobSpec creates a new job specification
// by passing the nomad job spec through a template
// engine with some values.
func (cl *Nomad) PrepareJobSpec(tempData map[string]interface{}) ([]byte, error) {
	temp, err := data.Asset("cocoon.job.json")
	if err != nil {
		return nil, err
	}

	str := util.RenderTemp(string(temp), tempData)
	return []byte(str), nil
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
func (sc *Nomad) Deploy(jobID, lang, url, tag, buildParams, memory, cpuShare string) (*DeploymentInfo, error) {

	var err error

	if len(jobID) == 0 {
		return nil, fmt.Errorf("job id is required")
	}

	if err = common.ValidateDeployment(url, lang, buildParams); err != nil {
		return nil, err
	}
	if !util.InStringSlice(util.GetMapKeys(SupportedMemory), memory) {
		return nil, fmt.Errorf("Invalid memory value. Expects one of these: %v", util.GetMapKeys(SupportedMemory))
	} else if !util.InStringSlice(util.GetMapKeys(SupportedCPUShare), cpuShare) {
		return nil, fmt.Errorf("Invalid cpu share value. Expects one of these: %v", util.GetMapKeys(SupportedCPUShare))
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

	cocoonData := map[string]interface{}{
		"ID":                jobID,
		"Count":             1,
		"MemoryMB":          SupportedMemory[memory].(int),
		"CPU":               SupportedCPUShare[cpuShare].(int),
		"DiskMB":            SupportedDiskSpace[cpuShare].(int),
		"MBits":             1000,
		"Image":             img,
		"CocoonCodeURL":     url,
		"CocoonCodeLang":    lang,
		"CocoonCodeTag":     tag,
		"CocoonBuildParams": buildParams,
	}

	jobSpec, err := sc.PrepareJobSpec(cocoonData)

	if err != nil {
		return nil, fmt.Errorf("system: failed to prepare job spec. %s", err)
	}

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
