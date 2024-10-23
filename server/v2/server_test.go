package serverv2_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	coreserver "cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	grpc "cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/store"
	storev2 "cosmossdk.io/store/v2"
)

type mockInterfaceRegistry struct{}

func (*mockInterfaceRegistry) Resolve(typeUrl string) (gogoproto.Message, error) {
	panic("not implemented")
}

func (*mockInterfaceRegistry) ListImplementations(ifaceTypeURL string) []string {
	panic("not implemented")
}
func (*mockInterfaceRegistry) ListAllInterfaces() []string { panic("not implemented") }

type mockApp[T transaction.Tx] struct {
	serverv2.AppI[T]
}

func (*mockApp[T]) QueryHandlers() map[string]appmodulev2.Handler {
	return map[string]appmodulev2.Handler{}
}

func (*mockApp[T]) InterfaceRegistry() coreserver.InterfaceRegistry {
	return &mockInterfaceRegistry{}
}

func (*mockApp[T]) Store() storev2.RootStore {
	return nil
}

func TestServer(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	configPath := filepath.Join(currentDir, "testdata")

	v, err := serverv2.ReadConfig(configPath)
	if err != nil {
		v = viper.New()
	}
	cfg := v.AllSettings()

	logger := log.NewLogger(os.Stdout)

	ctx, err := serverv2.SetServerContext(context.Background(), v, logger)
	require.NoError(t, err)

	grpcServer := grpc.New[transaction.Tx]()
	err = grpcServer.Init(&mockApp[transaction.Tx]{}, cfg, logger)
	require.NoError(t, err)

	storeServer := store.New[transaction.Tx]()
	err = storeServer.Init(&mockApp[transaction.Tx]{}, cfg, logger)
	require.NoError(t, err)

	mockServer := &mockServer{name: "mock-server-1", ch: make(chan string, 100)}

	server := serverv2.NewServer(
		serverv2.DefaultServerConfig(),
		grpcServer,
		storeServer,
		mockServer,
	)

	serverCfgs := server.Configs()
	require.Equal(t, serverCfgs[grpcServer.Name()].(*grpc.Config).Address, grpc.DefaultConfig().Address)
	require.Equal(t, serverCfgs[mockServer.Name()].(*mockServerConfig).MockFieldOne, MockServerDefaultConfig().MockFieldOne)

	// write config
	err = server.WriteConfig(configPath)
	require.NoError(t, err)

	v, err = serverv2.ReadConfig(configPath)
	require.NoError(t, err)

	require.Equal(t, v.GetString(grpcServer.Name()+".address"), grpc.DefaultConfig().Address)

	// start empty
	ctx, cancelFn := context.WithCancel(ctx)
	go func() {
		// wait 5sec and cancel context
		<-time.After(5 * time.Second)
		cancelFn()

		err = server.Stop(ctx)
		require.NoError(t, err)
	}()

	err = server.Start(ctx)
	require.NoError(t, err)
}
