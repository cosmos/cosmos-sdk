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

func NewMockServer(name string) *MockServer {
	return &MockServer{
		name: name,
		ch:   make(chan string, 100),
	}
}

func (s *MockServer) Start(ctx context.Context) error {
	for ctx.Err() == nil {
		s.ch <- fmt.Sprintf("%s mock server: %d", s.name, rand.Int())
	}

	close(s.ch)
	return nil
}

func (s *MockServer) Stop(ctx context.Context) error {
	for range s.ch {
		if str := <-s.ch; str != "" {
			fmt.Printf("Clearing %s\n", str)
		}
	}

	return nil
}
