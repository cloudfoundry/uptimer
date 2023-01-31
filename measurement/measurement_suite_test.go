package measurement_test

import (
	"bytes"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMeasurement(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Measurement Suite")
}

// A buffer is a thread-safe wrapper around bytes.Buffer.
type buffer struct {
	b  *bytes.Buffer
	mu sync.RWMutex
}

// newBuffer creates a buffer.
func newBuffer() *buffer {
	return &buffer{b: bytes.NewBuffer([]byte{})}
}

// Write appends the contents of p to the buffer in a thread-safe way, growing
// the buffer as needed. The return value n is the length of p; err is always
// nil. If the buffer becomes too large, Write will panic with ErrTooLarge.
func (b *buffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.b.Write(p)
}

// string returns the contents of the unread portion of the buffer as a string
// in a thread-safe way. If the Buffer is a nil pointer, it returns "<nil>".
func (b *buffer) string() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.b.String()
}
