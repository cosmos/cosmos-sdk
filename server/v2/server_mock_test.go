package serverv2_test

import (
	"context"
	"fmt"
	"math/rand"
)

type mockServer struct {
	name string
	ch   chan string
}

func NewServer(name string) *mockServer {
	return &mockServer{
		name: name,
		ch:   make(chan string, 100),
	}
}

func (s *mockServer) Name() string {
	return s.name
}

func (s *mockServer) Start(ctx context.Context) error {
	for ctx.Err() == nil {
		s.ch <- fmt.Sprintf("%s mock server: %d", s.name, rand.Int())
	}

	return nil
}

func (s *mockServer) Stop(ctx context.Context) error {
	for range s.ch {
		if str := <-s.ch; str != "" {
			fmt.Printf("clearing %s\n", str)
		}
	}

	return nil
}

func (s *mockServer) Config() interface{} {
	return struct {
		MockFieldOne string `mapstructure:"mock_field" toml:"mock_field" comment:"Mock field"`
		MockFieldTwo int    `mapstructure:"mock_field_two" toml:"mock_field_two" comment:"Mock field two"`
	}{
		MockFieldOne: "mock",
		MockFieldTwo: 2,
	}
}
