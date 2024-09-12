package serverv2_test

import (
	"context"
	"fmt"
	"math/rand"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
)

type mockServerConfig struct {
	MockFieldOne string `mapstructure:"mock_field" toml:"mock_field" comment:"Mock field"`
	MockFieldTwo int    `mapstructure:"mock_field_two" toml:"mock_field_two" comment:"Mock field two"`
}

func MockServerDefaultConfig() *mockServerConfig {
	return &mockServerConfig{
		MockFieldOne: "default",
		MockFieldTwo: 1,
	}
}

type mockServer struct {
	name string
	ch   chan string
}

func (s *mockServer) Name() string {
	return s.name
}

func (s *mockServer) Init(appI serverv2.AppI[transaction.Tx], cfg map[string]any, logger log.Logger) error {
	return nil
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

func (s *mockServer) Config() any {
	return MockServerDefaultConfig()
}
