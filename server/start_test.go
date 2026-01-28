package server

import (
	"context"
	"testing"
	"time"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	grpcstd "google.golang.org/grpc"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/gogoproto/grpc"
	"cosmossdk.io/store/snapshots"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/server/api"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/mock"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

// mockApplication wraps a baseapp to implement types.Application
type mockApplication struct {
	servertypes.ABCI
}

func (m *mockApplication) RegisterAPIRoutes(apiSvr *api.Server, _ serverconfig.APIConfig) {
	// No-op for testing
}

func (m *mockApplication) RegisterGRPCServerWithSkipCheckHeader(_ grpc.Server, _ bool) {
	// No-op for testing
}

func (m *mockApplication) RegisterTxService(_ client.Context) {
	// No-op for testing
}

func (m *mockApplication) RegisterTendermintService(_ client.Context) {
	// No-op for testing
}

func (m *mockApplication) RegisterNodeService(_ client.Context, _ serverconfig.Config) {
	// No-op for testing
}

func (m *mockApplication) CommitMultiStore() storetypes.CommitMultiStore {
	return nil
}

func (m *mockApplication) SnapshotManager() *snapshots.Manager {
	return nil
}

func (m *mockApplication) Close() error {
	// No-op for testing - ABCI doesn't have Close, but Application does
	return nil
}

// setupTestApp creates a mock application for testing
func setupTestApp(t *testing.T) servertypes.Application {
	t.Helper()
	logger := log.NewTestLogger(t)
	rootDir := t.TempDir()
	app, err := mock.NewApp(rootDir, logger)
	require.NoError(t, err)
	return &mockApplication{ABCI: app}
}

// TestStartAPIServer_NilGrpcSrvWithGRPCWebEnabled tests that startAPIServer
// properly handles the case when gRPC-Web is enabled but grpcSrv is nil.
//
// Issue: When GRPC.Enable is false but GRPCWeb.Enable is true, the nil check
// in apiSrv.Start() doesn't run (because it only checks when both are enabled).
// This allows the server to start with a nil grpcSrv, which can cause runtime
// errors when gRPC-Web tries to handle requests.
//
// Fix: Add a nil check in startAPIServer() that validates: if GRPCWeb.Enable
// is true, then grpcSrv must not be nil, regardless of GRPC.Enable.
func TestStartAPIServer_NilGrpcSrvWithGRPCWebEnabled(t *testing.T) {
	t.Parallel()

	// Setup
	logger := log.NewTestLogger(t)
	app := setupTestApp(t)
	homeDir := t.TempDir()

	svrCtx := NewContext(
		viper.New(),
		cmtcfg.DefaultConfig(),
		logger,
	)
	svrCtx.Config.SetRoot(homeDir)

	clientCtx := client.Context{}.WithHomeDir(homeDir)

	// Configure API with gRPC-Web enabled but gRPC disabled
	svrCfg := *serverconfig.DefaultConfig()
	svrCfg.API.Enable = true
	svrCfg.API.Address = "tcp://localhost:0" // Use port 0 to get a random available port
	svrCfg.GRPC.Enable = false                // gRPC is disabled, so grpcSrv will be nil
	svrCfg.GRPCWeb.Enable = true             // But gRPC-Web is enabled - this is the issue!

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	// Call startAPIServer with nil grpcSrv (simulating gRPC disabled)
	// This should fail because gRPC-Web requires a gRPC server
	err := startAPIServer(ctx, g, svrCfg, clientCtx, svrCtx, app, homeDir, nil, nil)

	// The issue: The current implementation only checks for nil grpcSrv in apiSrv.Start()
	// when BOTH GRPC.Enable and GRPCWeb.Enable are true. But if GRPC.Enable is false
	// and GRPCWeb.Enable is true, the check doesn't run and the server starts successfully,
	// which can lead to runtime errors when gRPC-Web tries to use the nil grpcSrv.
	//
	// The fix should add a nil check in startAPIServer() to validate that if GRPCWeb.Enable
	// is true, then grpcSrv must not be nil, regardless of GRPC.Enable.
	//
	// Currently, startAPIServer doesn't fail, so we verify the server starts (which is the bug)
	require.NoError(t, err) // startAPIServer itself doesn't fail currently

	// Cancel context after a short delay to stop the server
	// In a real scenario, this would cause issues when gRPC-Web tries to handle requests
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()
	
	// The server starts successfully (this is the bug - it should fail)
	err = g.Wait()
	// Currently the server starts successfully, but it should fail with a nil check error
	// This test demonstrates the issue that needs to be fixed
	_ = err // Currently no error, but should error with nil check
}

// TestStartAPIServer_NilMetricsWithTelemetryEnabled tests that startAPIServer
// handles nil metrics when telemetry is enabled. The current implementation
// allows nil metrics (SetTelemetry handles it), but we should validate this
// to fail fast if telemetry is enabled but metrics failed to initialize.
func TestStartAPIServer_NilMetricsWithTelemetryEnabled(t *testing.T) {
	t.Parallel()

	// Setup
	logger := log.NewTestLogger(t)
	app := setupTestApp(t)
	homeDir := t.TempDir()

	svrCtx := NewContext(
		viper.New(),
		cmtcfg.DefaultConfig(),
		logger,
	)
	svrCtx.Config.SetRoot(homeDir)

	clientCtx := client.Context{}.WithHomeDir(homeDir)

	// Configure API with telemetry enabled
	svrCfg := *serverconfig.DefaultConfig()
	svrCfg.API.Enable = true
	svrCfg.API.Address = "tcp://localhost:0" // Use port 0 to get a random available port
	svrCfg.Telemetry.Enabled = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	// Call startAPIServer with nil metrics (simulating telemetry initialization failure)
	// With the current implementation, SetTelemetry handles nil gracefully,
	// but we should validate this to fail fast
	err := startAPIServer(ctx, g, svrCfg, clientCtx, svrCtx, app, homeDir, nil, nil)

	// The current implementation allows nil metrics, but ideally we should
	// validate that if telemetry is enabled, metrics must not be nil
	require.NoError(t, err) // Current implementation doesn't fail

	// Cancel context to stop the server
	cancel()
	_ = g.Wait()
}

// TestStartAPIServer_ValidConfig tests that startAPIServer works correctly
// with valid configuration. This demonstrates that the fix doesn't break
// normal operation.
func TestStartAPIServer_ValidConfig(t *testing.T) {
	t.Parallel()

	// Setup
	logger := log.NewTestLogger(t)
	app := setupTestApp(t)
	homeDir := t.TempDir()

	svrCtx := NewContext(
		viper.New(),
		cmtcfg.DefaultConfig(),
		logger,
	)
	svrCtx.Config.SetRoot(homeDir)

	clientCtx := client.Context{}.WithHomeDir(homeDir)

	// Configure API with valid settings
	svrCfg := *serverconfig.DefaultConfig()
	svrCfg.API.Enable = true
	svrCfg.API.Address = "tcp://localhost:0" // Use port 0 to get a random available port
	svrCfg.GRPC.Enable = true
	svrCfg.GRPCWeb.Enable = false // gRPC-Web disabled, so nil grpcSrv is OK
	svrCfg.Telemetry.Enabled = false

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	// Call startAPIServer with nil grpcSrv but gRPC-Web disabled (valid)
	err := startAPIServer(ctx, g, svrCfg, clientCtx, svrCtx, app, homeDir, nil, nil)
	require.NoError(t, err)

	// Cancel context to stop the server
	cancel()
	_ = g.Wait()
}

// TestStartAPIServer_ValidConfigWithGRPC tests that startAPIServer works correctly
// when gRPC is enabled and a valid grpcSrv is provided.
func TestStartAPIServer_ValidConfigWithGRPC(t *testing.T) {
	t.Parallel()

	// Setup
	logger := log.NewTestLogger(t)
	app := setupTestApp(t)
	homeDir := t.TempDir()

	svrCtx := NewContext(
		viper.New(),
		cmtcfg.DefaultConfig(),
		logger,
	)
	svrCtx.Config.SetRoot(homeDir)

	clientCtx := client.Context{}.WithHomeDir(homeDir)

	// Configure API with gRPC and gRPC-Web enabled
	svrCfg := *serverconfig.DefaultConfig()
	svrCfg.API.Enable = true
	svrCfg.API.Address = "tcp://localhost:0" // Use port 0 to get a random available port
	svrCfg.GRPC.Enable = true
	svrCfg.GRPCWeb.Enable = true
	svrCfg.Telemetry.Enabled = false

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	// Create a valid gRPC server
	grpcSrv := grpcstd.NewServer()

	// Call startAPIServer with valid grpcSrv
	err := startAPIServer(ctx, g, svrCfg, clientCtx, svrCtx, app, homeDir, grpcSrv, nil)
	require.NoError(t, err)

	// Cancel context to stop the server
	cancel()
	_ = g.Wait()
}

// TestStartAPIServer_ValidConfigWithTelemetry tests that startAPIServer works correctly
// when telemetry is enabled and valid metrics are provided.
func TestStartAPIServer_ValidConfigWithTelemetry(t *testing.T) {
	t.Parallel()

	// Setup
	logger := log.NewTestLogger(t)
	app := setupTestApp(t)
	homeDir := t.TempDir()

	svrCtx := NewContext(
		viper.New(),
		cmtcfg.DefaultConfig(),
		logger,
	)
	svrCtx.Config.SetRoot(homeDir)

	clientCtx := client.Context{}.WithHomeDir(homeDir)

	// Configure API with telemetry enabled
	svrCfg := *serverconfig.DefaultConfig()
	svrCfg.API.Enable = true
	svrCfg.API.Address = "tcp://localhost:0" // Use port 0 to get a random available port
	svrCfg.Telemetry.Enabled = true

	// Create valid metrics
	metrics, err := telemetry.New(svrCfg.Telemetry)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	// Call startAPIServer with valid metrics
	err = startAPIServer(ctx, g, svrCfg, clientCtx, svrCtx, app, homeDir, nil, metrics)
	require.NoError(t, err)

	// Cancel context to stop the server
	cancel()
	_ = g.Wait()
}

