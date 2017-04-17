package client

import (
	"context"
	"fmt"
	"time"

	"github.com/ellcrys/util"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MinimumLogLines defines the minimum number of lines of logs to return
var MinimumLogLines = 10

// GetLogs displays the logs of a cocoon. Supports
// number of line limitation via the numLines parameter.
// Continuous streaming of logs if tail is true and filter
// by stderr or stdout. By default stderr and stdout are returned.
func GetLogs(cocoonID string, numLines int, tail, stderrOnly, stdoutOnly bool) error {

	if numLines < MinimumLogLines {
		numLines = MinimumLogLines
	}

	source := ""
	if stderrOnly {
		source = "stderr"
	}
	if stdoutOnly {
		source = "stdout"
	}

	userSession, err := GetUserSessionToken()
	if err != nil {
		return err
	}

	// stopSpinner := util.Spinner("Please wait")

	conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("unable to connect to cluster. please try again")
	}
	defer conn.Close()

	ctx := metadata.NewContext(context.Background(), metadata.Pairs("access_token", userSession.Token))
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	cl := proto.NewAPIClient(conn)
	resp, err := cl.GetLogs(ctx, &proto.GetLogsRequest{
		CocoonID:   cocoonID,
		NumLines:   int32(numLines),
		FollowTail: tail,
		Source:     source,
	})
	if err != nil {
		return err
	}

	var messages []types.LogMessage
	util.FromJSON(resp.Body, &messages)

	for i := len(messages) - 1; i >= 0; i-- {
		// fmt.Println(fmt.Sprintf(color.Cyan("(%s)")+": %s", cocoonID, messages[i].Text))
	}

	return nil
}
