package serverv2_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	grpc "cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/core/appmanager"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

type grpcClientCtx struct {
	interfaceRegistry codectypes.InterfaceRegistry
}

func (c *grpcClientCtx) InterfaceRegistry() appmanager.InterfaceRegistry {
	return c.interfaceRegistry
}

func TestServer(t *testing.T) {
	log := log.NewLogger(os.Stdout)
	interfaceRegistry := &grpcClientCtx{codectypes.NewInterfaceRegistry()}

	// TODO we need to have the server gets the latest config
	grpcServer, err := grpc.New(interfaceRegistry, log, nil)
	if err != nil {
		panic(err)
	}

	server := serverv2.NewServer(
		log,
		grpcServer,
	)

	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	configPath := filepath.Join(currentDir, "app.toml")

	// write config
	if err := server.WriteConfig(configPath); err != nil {
		panic(err)
	}

	// read config
	v := server.Config(configPath)
	fmt.Println(v)
}
