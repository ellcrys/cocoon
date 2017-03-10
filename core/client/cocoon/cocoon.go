package cocoon

import (
	"fmt"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/ellcrys/util"
	cocoon_util "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/api/grpc/proto"
	"github.com/ncodes/cocoon/core/config"
	"github.com/ncodes/cocoon/core/types/client"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("api.client")

// APIAddress is the remote address to the cluster server
var APIAddress = util.Env("API_ADDRESS", "127.0.0.1:8004")

// supportedLanguages list the languages supported
var supportedLanguages = []string{"go"}

// supportedMemory list the memory values supported
var supportedMemory = []string{"512m", "1g", "2g"}

// supportedCPUShare list the cpu share values supported
var supportedCPUShare = []string{"1x", "2x"}

func init() {
	log.SetBackend(config.MessageOnlyBackend)
}

// Ops defines cocoon operations
type Ops struct {
}

// validateCreateCocoon validates a cocoon to be created
func validateCreateCocoon(c *client.Cocoon) error {

	if len(c.URL) == 0 {
		return fmt.Errorf("url is required")
	} else if !cocoon_util.IsGithubRepoURL(c.URL) {
		return fmt.Errorf("url is not a valid github repo url")
	} else if len(c.Language) == 0 {
		return fmt.Errorf("language is required")
	} else if !util.InStringSlice(supportedLanguages, c.Language) {
		return fmt.Errorf("language is not supported. Expects one of these values %s", supportedLanguages)
	} else if len(c.BuildParam) > 0 {
		var _c map[string]interface{}
		if util.FromJSON([]byte(c.BuildParam), &_c) != nil {
			return fmt.Errorf("build parameter is not valid json")
		}
	} else if len(c.Memory) == 0 {
		return fmt.Errorf("memory is required")
	} else if !util.InStringSlice(supportedMemory, c.Memory) {
		return fmt.Errorf("Memory value is not supported. Expects one of these values %s", supportedMemory)
	} else if len(c.CPUShare) == 0 {
		return fmt.Errorf("CPU share is required")
	} else if !util.InStringSlice(supportedCPUShare, c.CPUShare) {
		return fmt.Errorf("CPU share value is not supported. Expects one of these values %s", supportedCPUShare)
	} else if c.Instances > 10 {
		return fmt.Errorf("Instances value is currently limited to 10")
	}

	return nil
}

// Create a new cocoon locally
func (c *Ops) Create(cocoon *client.Cocoon) error {

	id := util.UUID4()
	cocoon.ID = id

	err := validateCreateCocoon(cocoon)
	if err != nil {
		return err
	}

	release := client.Release{
		ID:         util.UUID4(),
		CocoonID:   cocoon.ID,
		URL:        cocoon.URL,
		ReleaseTag: cocoon.ReleaseTag,
		Language:   cocoon.Language,
		BuildParam: cocoon.BuildParam,
	}

	cocoon.Releases = []string{release.ID}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	defer stopSpinner()

	client := proto.NewAPIClient(conn)
	resp, err := client.CreateCocoon(context.Background(), &proto.CreateCocoonRequest{
		Id:             cocoon.ID,
		Url:            cocoon.URL,
		Language:       cocoon.Language,
		ReleaseTag:     cocoon.ReleaseTag,
		BuildParam:     []byte(cocoon.BuildParam),
		Memory:         cocoon.Memory,
		CPUShare:       cocoon.CPUShare,
		Releases:       cocoon.Releases,
		Instances:      cocoon.Instances,
		NumSignatories: cocoon.NumSignatories,
		SigThreshold:   cocoon.SigThreshold,
	})

	if err != nil {
		stopSpinner()
		return err
	} else if resp.Status != 200 {
		stopSpinner()
		return fmt.Errorf("%s", resp.Body)
	}

	resp, err = client.CreateRelease(context.Background(), &proto.CreateReleaseRequest{
		Id:         release.ID,
		CocoonID:   cocoon.ID,
		Url:        cocoon.URL,
		Language:   cocoon.Language,
		ReleaseTag: cocoon.ReleaseTag,
		BuildParam: []byte(cocoon.BuildParam),
	})

	if err != nil {
		stopSpinner()
		return err
	} else if resp.Status != 200 {
		stopSpinner()
		return fmt.Errorf("%s", resp.Body)
	}

	stopSpinner()
	log.Info("==> New cocoon created")
	log.Infof("==> Cocoon ID: %s", id)
	log.Infof("==> Release ID: %s", release.ID)

	return nil
}

// Deploy creates and sends a deploy request to the server
func (c *Ops) deploy(cocoon *client.Cocoon) error {

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	client := proto.NewAPIClient(conn)
	resp, err := client.Deploy(context.Background(), &proto.DeployRequest{
		Id:         cocoon.ID,
		Url:        cocoon.URL,
		Language:   cocoon.Language,
		ReleaseTag: cocoon.ReleaseTag,
		BuildParam: []byte(cocoon.BuildParam),
		Memory:     cocoon.Memory,
		CpuShare:   cocoon.CPUShare,
	})
	if err != nil {
		return err
	} else if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	return nil
}

// Start starts a new or stopped cocoon code
func (c *Ops) Start(id string) error {

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	stopSpinner := util.Spinner("Please wait")
	cl := proto.NewAPIClient(conn)
	resp, err := cl.GetCocoon(context.Background(), &proto.GetCocoonRequest{
		Id: id,
	})

	if err != nil {
		stopSpinner()
		return err
	} else if resp.Status != 200 {
		stopSpinner()
		return fmt.Errorf("%s", resp.Body)
	}

	var cocoon client.Cocoon
	err = util.FromJSON(resp.Body, &cocoon)

	if err = c.deploy(&cocoon); err != nil {
		stopSpinner()
		return err
	}

	stopSpinner()
	log.Info("==> Successfully created a deployment request")
	log.Info("==> ID:", cocoon.ID)

	return nil
}
