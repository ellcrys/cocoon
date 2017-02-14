package pod

import (
	"fmt"

	"github.com/ncodes/cocoon/core/cluster"
	logging "github.com/op/go-logging"
)

var log = logging.MustGetLogger("deploy")

// Deploy creates a new pod on the cluster
// to run the smart contract app.
func Deploy(cluster cluster.Cluster, lang, url, tag string) (string, error) {
	log.Info(fmt.Sprintf("Deploying app with language=%s and url=%s", lang, url))
	return cluster.Deploy(lang, url, tag)
}
