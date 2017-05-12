package modo

import (
	"io"
	"sync"
)

// BufferSize is the amount of bytes to read from the writer every time Read() is called
var BufferSize = 256

// Outputter defines a structure for collecting
// data into a writer, reading this writer and sending the data
// to a callback function.
type Outputter struct {
	writer   io.Writer
	stopRead bool
	cb       func([]byte)
	m        sync.Mutex
}

// NewOutputter creates a new outputter
func NewOutputter(cb func([]byte)) *Outputter {
	return &Outputter{
		writer: NewSafeBuffer(),
		cb:     cb,
		m:      sync.Mutex{},
	}
}

// GetWriter returns a writer
func (o *Outputter) GetWriter() io.Writer {
	return o.writer
}

// Stop sets the stop read flag to true which effectively stops the
// the writer from being read from.
func (o *Outputter) Stop() {
	o.stopRead = true
}

// Start continuously reads the writer and sends the data to the output callback
func (o *Outputter) Start() error {
	tmp := make([]byte, BufferSize)
	buf := o.writer.(*SafeBuffer)
	for o.stopRead == false {
		if buf.Len() == 0 {
			continue
		}
		n, err := buf.Read(tmp)
		if err != nil {
			if err != io.EOF {
				return err
			}
			continue
		}
		if n > 0 {
			o.cb(tmp[:n])
		}
	}
	return nil
}
