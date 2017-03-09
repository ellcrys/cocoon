package cocoon

import (
	"fmt"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/grpc/proto"
	"github.com/ncodes/cocoon/core/common"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// var log = logging.MustGetLogger("api.client")

// APIAddress is the remote address to the cluster server
// var APIAddress = util.Env("API_ADDRESS", "127.0.0.1:8004")

// Deploy defines methods for deploying
// creating a deploy request.
type Deploy struct {
}

// Deploy creates and sends a deploy request to the server
func (cd *Deploy) Deploy(url, releaseTag, language, buildParam, memory, cpuShare string) error {

	if err := common.ValidateDeployment(url, language, buildParam); err != nil {
		return err
	}

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}

	id := util.UUID4()
	client := proto.NewAPIClient(conn)
	resp, err := client.Deploy(context.Background(), &proto.DeployRequest{
		Id:         id,
		Url:        url,
		Language:   language,
		ReleaseTag: releaseTag,
		BuildParam: []byte(buildParam),
		Memory:     memory,
		CpuShare:   cpuShare,
	})
	if err != nil {
		return err
	}

	if resp.Status != 200 {
		return fmt.Errorf("%s", resp.Body)
	}

	log.Info("==> Successfully deploy new cocoon")
	log.Info("==> ID:", id)

	return nil
}
