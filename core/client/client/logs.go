package client

import (
	"context"
	"fmt"
	"time"

	"strings"

	"github.com/ellcrys/util"
	"github.com/fatih/color"
	lru "github.com/hashicorp/golang-lru"
	"github.com/ncodes/cocoon/core/api/api/proto_api"
	"github.com/ncodes/cocoon/core/common"
	"github.com/ncodes/cocoon/core/types"
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

	conn, err := GetAPIConnection()
	if err != nil {
		return fmt.Errorf("unable to connect to the platform")
	}

	defer conn.Close()

	var fetch = func() (*proto_api.Response, error) {
		ctx, cc := context.WithTimeout(context.Background(), ContextTimeout)
		defer cc()
		cl := proto_api.NewAPIClient(conn)
		resp, err := cl.GetLogs(ctx, &proto_api.GetLogsRequest{
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
	seen, err := lru.New(10000)
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
			time.Sleep(100 * time.Millisecond)
			fmt.Println(
				fmt.Sprintf(color.CyanString("(%s):"), cocoonID),
				c.SprintfFunc()("%s", strings.TrimSpace(msg)),
			)
		}

		if !tail {
			return nil
		}

		time.Sleep(1 * time.Second)
		resp, err := fetch()
		if err != nil {
			return err
		}

		util.FromJSON(resp.Body, &messages)
	}
}
