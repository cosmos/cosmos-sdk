package serverv2_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	grpc "cosmossdk.io/server/v2/api/grpc"
)

type mockGRPCService struct {
	grpc.GRPCService
}

func (m *mockGRPCService) RegisterGRPCServer(gogogrpc.Server) {}

type mockInterfaceRegistry struct{}

func (*mockInterfaceRegistry) Resolve(typeUrl string) (gogoproto.Message, error) {
	panic("not implemented")
}

func (*mockInterfaceRegistry) ListImplementations(ifaceTypeURL string) []string {
	panic("not implemented")
}
func (*mockInterfaceRegistry) ListAllInterfaces() []string { panic("not implemented") }

// TODO split this test into multiple tests
// test read config
// test write config
// test server configs
// test start empty
// test start config exists
// test stop
func TestServer(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	configPath := filepath.Join(currentDir, "testdata")

	v, err := serverv2.ReadConfig(configPath)
	if err != nil {
		v = viper.New()
	}

	logger := log.NewLogger(os.Stdout)
	grpcServer, err := grpc.New(logger, v, &mockInterfaceRegistry{})
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

	serverCfgs := server.Configs()
	if serverCfgs[grpcServer.Name()].(*grpc.Config).Address != grpc.DefaultConfig().Address {
		t.Logf("config is not equal: %v", serverCfgs[grpcServer.Name()])
		t.Fail()
	}
	if serverCfgs[mockServer.Name()].(*mockServerConfig).MockFieldOne != MockServerDefaultConfig().MockFieldOne {
		t.Logf("config is not equal: %v", serverCfgs[mockServer.Name()])
		t.Fail()
	}

	// write config
	if err := server.WriteConfig(configPath); err != nil {
		t.Log(err)
		t.Fail()
	}

	v, err = serverv2.ReadConfig(configPath)
	if err != nil {
		t.Log(err) // config should be created by WriteConfig
		t.FailNow()
	}
	if v.GetString(grpcServer.Name()+".address") != grpc.DefaultConfig().Address {
		t.Logf("config is not equal: %v", v)
		t.Fail()
	}

	// start empty
	ctx := context.Background()
	ctx = context.WithValue(ctx, serverv2.ServerContextKey, serverv2.Config{StartBlock: true})
	ctx, cancelFn := context.WithCancel(ctx)
	go func() {
		// wait 5sec and cancel context
		<-time.After(5 * time.Second)
		cancelFn()

		if err := server.Stop(ctx); err != nil {
			t.Logf("failed to stop servers: %s", err)
			t.Fail()
		}
	}()

	if err := server.Start(ctx); err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestReadConfig(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	configPath := filepath.Join(currentDir, "testdata")

	v, err := serverv2.ReadConfig(configPath)
	require.NoError(t, err)

	grpcConfig := grpc.DefaultConfig()
	err = v.Sub("grpc-server").Unmarshal(&grpcConfig)
	require.NoError(t, err)
}
