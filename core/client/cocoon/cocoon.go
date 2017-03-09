package cocoon

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/ellcrys/util"
	cocoon_util "github.com/ncodes/cocoon-util"
	"github.com/ncodes/cocoon/core/client/db"
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
	} else if len(c.Lang) == 0 {
		return fmt.Errorf("language is required")
	} else if !util.InStringSlice(supportedLanguages, c.Lang) {
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
	db := db.GetDefaultDB()

	err := validateCreateCocoon(cocoon)
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("cocoons"))
		if err != nil {
			return fmt.Errorf("failed to create bucket. %s", err)
		}

		val, _ := util.ToJSON(cocoon)
		err = b.Put([]byte(id), val)
		if err != nil {
			return fmt.Errorf("failed to create cocoon. %s", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	log.Info("==> New cocoon created")
	log.Infof("==> ID: %s", id)
	return nil
}

// Start starts a new or stopped cocoon code
func (c *Ops) Start(cocoon *client.Cocoon) error {
	return nil
}
