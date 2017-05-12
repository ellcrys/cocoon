package modo

import (
	"bytes"
	"sync"
)

// SafeBuffer wraps a buffer and provides thread safe
// read and write operations
type SafeBuffer struct {
	b *bytes.Buffer
	m sync.RWMutex
}

// NewSafeBuffer creates a SafeBuffer
func NewSafeBuffer() *SafeBuffer {
	return &SafeBuffer{
		b: bytes.NewBuffer(nil),
		m: sync.RWMutex{},
	}
}

func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.Read(p)
}

func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

func (b *SafeBuffer) String() string {
	b.m.RLock()
	defer b.m.RUnlock()
	return b.b.String()
}

// Len returns the length of the buffer
func (b *SafeBuffer) Len() int {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Len()
}
