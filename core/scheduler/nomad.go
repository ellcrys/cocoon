package scheduler

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/franela/goreq"
	"github.com/ncodes/cocoon/core/data"
	"github.com/ncodes/cocoon/core/validation"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("nomad")

// SupportedCocoonCodeLang defines the supported chaincode language
var SupportedCocoonCodeLang = []string{"go"}

// Nomad defines a nomad cluster that implements
// cluster.Cluster interface. Every interaction with
// the cluster is handled here.
type Nomad struct {
	clusterAddr string
	API         string
}

// NewNomad creates a nomad cluster object
func NewNomad() *Nomad {
	return new(Nomad)
}

// SetAddr sets the nomad's API endpoint
func (cl *Nomad) SetAddr(addr string, https bool) {
	scheme := "http://"
	if https {
		scheme = "https://"
	}
	cl.API = scheme + addr
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
func (cl *Nomad) deployJob(jobSpec string) (string, int, error) {

	res, err := goreq.Request{
		Method: "POST",
		Uri:    cl.API + "/v1/jobs",
		Body:   jobSpec,
	}.Do()

	if err != nil {
		return "", 0, err
	}

	respStr, _ := res.Body.ToString()
	return respStr, res.StatusCode, nil
}

// Deploy a cocoon code to the cluster
func (cl *Nomad) Deploy(jobID, lang, url, tag, buildParams string) (*DeploymentInfo, error) {

	var err error

	if len(jobID) == 0 {
		return nil, fmt.Errorf("job id is required")
	}

	if err = validation.ValidateDeployment(url, lang, buildParams); err != nil {
		return nil, err
	}

	log.Debugf("Deploying cocoon code with language=%s, url=%s, tag=%s", lang, url, tag)

	if len(buildParams) > 0 {
		bs, _ := util.ToJSON([]byte(buildParams))
		buildParams = string(bs)
	}

	var img string
	switch lang {
	case "go":
		img = "ncodes/cocoon-launcher:latest"
	}

	cocoonData := map[string]interface{}{
		"ID":                jobID,
		"Count":             1,
		"CPU":               500,
		"MemoryMB":          512,
		"DiskMB":            100,
		"Image":             img,
		"CocoonCodeURL":     url,
		"CocoonCodeLang":    lang,
		"CocoonCodeTag":     tag,
		"CocoonBuildParams": buildParams,
	}

	jobSpec, err := cl.PrepareJobSpec(cocoonData)

	if err != nil {
		return nil, fmt.Errorf("system: failed to prepare job spec. %s", err)
	}

	resp, status, err := cl.deployJob(string(jobSpec))
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
