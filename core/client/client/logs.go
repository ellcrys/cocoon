package client

import (
	"context"
	"fmt"
	"time"

	"strings"

	"github.com/ellcrys/util"
	"github.com/fatih/color"
	lru "github.com/hashicorp/golang-lru"
	"github.com/ncodes/cocoon/core/api/api/proto"
	"github.com/ncodes/cocoon/core/common"
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
func GetLogs(cocoonID string, numLines int, tail, stderrOnly, stdoutOnly, disableColor bool) error {

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

	var fetch = func() (*proto.Response, error) {
		conn, err := grpc.Dial(APIAddress, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("unable to connect to cluster. please try again")
		}
		defer conn.Close()

		ctx := metadata.NewContext(context.Background(), metadata.Pairs("access_token", userSession.Token))
		ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
		defer cancel()
		cl := proto.NewAPIClient(conn)
		resp, err := cl.GetLogs(ctx, &proto.GetLogsRequest{
			CocoonID: cocoonID,
			NumLines: int32(numLines),
			Source:   source,
		})
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	resp, err := fetch()
	if err != nil {
		return err
	}

	var messages []types.LogMessage
	util.FromJSON(resp.Body, &messages)
	seen, err := lru.New(1000)
	if err != nil {
		return fmt.Errorf("failed to create log cache")
	}

	for {
		for i := len(messages) - 1; i >= 0; i-- {

			if seen.Contains(messages[i].ID) {
				continue
			} else {
				seen.Add(messages[i].ID, struct{}{})
			}

			c := color.New(color.FgHiWhite)
			msg := messages[i].Text
			if disableColor {
				msg = string(common.RemoveASCIIColors([]byte(msg)))
			}
			time.Sleep(300 * time.Millisecond)
			fmt.Println(
				fmt.Sprintf(color.CyanString("(%s):"), cocoonID),
				c.SprintfFunc()("%s", strings.TrimSpace(msg)),
			)
		}

		if !tail {
			return nil
		}

		time.Sleep(300 * time.Millisecond)
		resp, err := fetch()
		if err != nil {
			return err
		}

		util.FromJSON(resp.Body, &messages)
	}
}
