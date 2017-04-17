package api

import (
	"fmt"

	"cloud.google.com/go/logging/logadmin"
	ptype "github.com/golang/protobuf/ptypes/struct"
	"github.com/ncodes/cocoon/core/types"
	context "golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// StackDriverLog implements LogProvider interface to
// allow the retrieval of logs from GCP stack driver.
type StackDriverLog struct {
	projectID string
	client    *logadmin.Client
}

// Init initializes the instance
func (s *StackDriverLog) Init(config map[string]interface{}) error {
	var err error
	var ok bool

	if s.projectID, ok = config["projectId"].(string); !ok {
		return fmt.Errorf("project id is required")
	}

	ctx := context.Background()
	s.client, err = logadmin.NewClient(ctx, s.projectID)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	return nil
}

// Get returns a slice of log messages. It will return the a maximum of recent
// numEntries entries. If source is not set, both stderr and stdout errors will be returned
func (s *StackDriverLog) Get(ctx context.Context, logName string, numEntries int, source string) ([]types.LogMessage, error) {

	if s.client == nil {
		return nil, fmt.Errorf("client not initialized. Did you call Init()?")
	}

	opts := []logadmin.EntriesOption{
		logadmin.Filter(fmt.Sprintf(`logName = "projects/%s/logs/%s"`, s.projectID, logName)),
		logadmin.NewestFirst(),
	}
	if len(source) == 0 {
		opts = append(opts, logadmin.Filter(`jsonPayload.source="stderr" OR "stdout"`))
	} else {
		if source == "stderr" || source == "stdout" {
			opts = append(opts, logadmin.Filter(fmt.Sprintf(`jsonPayload.source="%s"`, source)))
		} else {
			return nil, fmt.Errorf("invalid source: %s", source)
		}
	}

	var messages []types.LogMessage
	iter := s.client.Entries(ctx, opts...)
	for len(messages) < numEntries {
		entry, err := iter.Next()
		if err == iterator.Done {
			return messages, nil
		}
		if err != nil {
			return nil, err
		}

		if payload, ok := entry.Payload.(*ptype.Struct); ok {
			messages = append(messages, types.LogMessage{
				ID:        entry.InsertID,
				Text:      payload.GetFields()["log"].GetStringValue(),
				Timestamp: entry.Timestamp,
			})
		}
	}

	return messages, nil
}
