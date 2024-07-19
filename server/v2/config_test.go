package serverv2_test

import (
	"os"
	"path/filepath"
	"testing"

	serverv2 "cosmossdk.io/server/v2"
	grpc "cosmossdk.io/server/v2/api/grpc"
	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	configPath := filepath.Join(currentDir, "testdata")

	v, err := serverv2.ReadConfig(configPath)
	require.NoError(t, err)

	require.Equal(t, v.GetString("grpc.address"), grpc.DefaultConfig().Address)
}

func TestUnmarshalSubConfig(t *testing.T) {
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	configPath := filepath.Join(currentDir, "testdata")

	v, err := serverv2.ReadConfig(configPath)
	require.NoError(t, err)

	grpcConfig := grpc.DefaultConfig()
	err = serverv2.UnmarshalSubConfig(v, "grpc", &grpcConfig)
	require.NoError(t, err)

	require.True(t, grpc.DefaultConfig().Enable)
	require.False(t, grpcConfig.Enable)
}
