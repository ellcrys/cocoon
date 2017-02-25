package cluster

import (
	"fmt"

	"strings"

	"github.com/ellcrys/util"
	"github.com/franela/goreq"
	cutil "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/data"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("nomad")

// SupportedChaincodeLang defines the supported chaincode language
var SupportedChaincodeLang = []string{"go"}

// Nomad defines a nomad cluster that implements
// cluster.Cluster interface. Every interaction with
// the cluster is handled here.
type Nomad struct {
	clusterAddr string
	API         string
}

// NewNomad creates a nomad cluster object
func NewNomad() *Nomad {
	log.Info("Nomad cluster instance created")
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

// Deploy a smart contract to the cluster
func (cl *Nomad) Deploy(lang, url, tag, buildParams string) (string, error) {

	if !util.InStringSlice(SupportedChaincodeLang, lang) {
		return "", fmt.Errorf("only the following languages are suppored [%s]", strings.Join(SupportedChaincodeLang, ","))
	} else if url == "" {
		return "", fmt.Errorf("github repo url is required")
	} else if !cutil.IsGithubRepoURL(url) {
		return "", fmt.Errorf("invalid chaincode url. Chaincode must be hosted on github")
	}

	var img string
	switch lang {
	case "go":
		img = "ncodes/cocoon-launcher:latest"
	}

	cocoonData := map[string]interface{}{
		"ID":                util.Sha1(util.UUID4())[0:15],
		"Count":             1,
		"CPU":               500,
		"MemoryMB":          512,
		"DiskMB":            300,
		"Image":             img,
		"CocoonCodeURL":     url,
		"CocoonCodeLang":    lang,
		"CocoonCodeTag":     tag,
		"CocoonBuildParams": buildParams,
	}

	jobSpec, err := cl.PrepareJobSpec(cocoonData)

	if err != nil {
		e := fmt.Errorf("failed to prepare job spec. %s", err)
		log.Error(e.Error())
		return "", e
	}

	resp, status, err := cl.deployJob(string(jobSpec))
	if err != nil {
		e := fmt.Errorf("failed to deploy job spec. %s", err)
		log.Error(e.Error())
		return "", e
	} else if status != 200 {
		e := fmt.Errorf("failed to deploy job spec. %s", resp)
		log.Error(resp)
		return "", e
	}

	return cocoonData["ID"].(string), nil
}
