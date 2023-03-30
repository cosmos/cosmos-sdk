package cosmovisor_test

import (
	"bytes"
	"sync"
)

// buffer is a thread safe bytes buffer
type Buffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func NewBuffer() *Buffer {
	return &Buffer{}
}

func (b *Buffer) Write(bz []byte) (int, error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(bz)
}

func (b *Buffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func (b *Buffer) Reset() {
	b.m.Lock()
	defer b.m.Unlock()
	b.b.Reset()
}
