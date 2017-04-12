package connector

import (
	"bytes"
	"io"

	"io/ioutil"

	"time"

	"sync"

	logging "github.com/op/go-logging"
)

var logStreamLogger = logging.MustGetLogger("stream_logger")

// LogStreamer defines a structure for collecting
// data into a writer and writing to a logger.
type LogStreamer struct {
	writer   io.Writer
	logger   *logging.Logger
	stopRead bool
	mutex    sync.Mutex
}

// NewLogStreamer reads from a stream and logs to stdout
func NewLogStreamer() *LogStreamer {
	return &LogStreamer{
		writer: bytes.NewBuffer(nil),
		logger: logStreamLogger,
		mutex:  sync.Mutex{},
	}
}

// GetWriter returns a writer
func (ls *LogStreamer) GetWriter() io.Writer {
	return ls.writer
}

// SetLogger sets the logger
func (ls *LogStreamer) SetLogger(logger *logging.Logger) {
	ls.logger = logger
}

// Stop sets the stop read flag to true which effectively stops the
// the writer from being read and logged.
func (ls *LogStreamer) Stop() {
	ls.stopRead = true
}

// Start starts reading the writer and logging its data
func (ls *LogStreamer) Start() error {
	for ls.stopRead == false {
		ls.mutex.Lock()
		bs, err := ioutil.ReadAll(ls.writer.(*bytes.Buffer))
		if err != nil {
			ls.mutex.Unlock()
			return err
		}

		// reset the writer
		ls.writer.(*bytes.Buffer).Reset()
		ls.mutex.Unlock()

		if len(bs) > 0 {
			ls.logger.Info(string(bs))
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
