package mock

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/serverv2"
)

type MockServer struct {
	name string
	ch   chan string
}

var _ serverv2.Module = (*MockServer)(nil)

func NewServer(name string) *MockServer {
	return &MockServer{
		name: name,
		ch:   make(chan string, 100),
	}
}

func (s *MockServer) Name() string {
	return s.name
}

func (s *MockServer) Start(ctx context.Context) error {
	for ctx.Err() == nil {
		s.ch <- fmt.Sprintf("%s mock server: %d", s.name, rand.Int())
	}

	return nil
}

func (s *MockServer) Stop(ctx context.Context) error {
	for range s.ch {
		if str := <-s.ch; str != "" {
			fmt.Printf("clearing %s\n", str)
		}
	}

	return nil
}
