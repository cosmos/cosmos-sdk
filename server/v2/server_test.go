package serverv2_test

import (
	"os"
	"path/filepath"
	"testing"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	grpc "cosmossdk.io/server/v2/api/grpc"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/spf13/viper"
)

type mockGRPCService struct {
	grpc.GRPCService
}

func (m *mockGRPCService) RegisterGRPCServer(gogogrpc.Server) {}

func TestServer(t *testing.T) {
	logger := log.NewLogger(os.Stdout)
	interfaceRegistry := codectypes.NewInterfaceRegistry()

	// TODO we need to have the server gets the latest config
	grpcServer, err := grpc.New(logger, viper.New(), interfaceRegistry, &mockGRPCService{})
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	mockServer := &mockServer{name: "mock-server-1", ch: make(chan string, 100)}

	server := serverv2.NewServer(
		logger,
		grpcServer,
		mockServer,
	)

	currentDir, err := os.Getwd()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	configPath := filepath.Join(currentDir, "app.toml")

	// write config
	if err := server.WriteConfig(configPath); err != nil {
		t.Log(err)
		t.Fail()
	}

	// read config
	v, err := server.Config(configPath)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if v == nil {
		t.Log("config is nil")
		t.FailNow()
	}

	if v.GetString(grpcServer.Name()+".address") != grpc.DefaultConfig().Address {
		t.Logf("config is not equal: %v", v)
		t.Fail()
	}
}
