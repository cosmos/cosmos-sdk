package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cosmossdk.io/log"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/cmdtest"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// ExportSystem wraps a (*cmdtest).System
// and sets up appropriate client and server contexts,
// to simplify testing the export CLI.
type ExportSystem struct {
	sys *cmdtest.System

	Ctx context.Context

	Sctx *server.Context
	Cctx client.Context

	HomeDir string
}

// newExportSystem returns a cmdtest.System with export as a child command,
// and it returns a context.Background with an associated *server.Context value.
func NewExportSystem(t *testing.T, exporter types.AppExporter) *ExportSystem {
	t.Helper()

	homeDir := t.TempDir()

	// Unclear why we have to create the config directory ourselves,
	// but tests fail without this.
	if err := os.MkdirAll(filepath.Join(homeDir, "config"), 0o700); err != nil {
		t.Fatal(err)
	}

	sys := cmdtest.NewSystem()
	sys.AddCommands(
		server.ExportCmd(exporter, homeDir),
		genutilcli.InitCmd(module.NewBasicManager(), homeDir),
	)

	tw := zerolog.NewTestWriter(t)
	tw.Frame = 5 // Seems to be the magic number to get source location to match logger calls.

	sCtx := server.NewContext(
		viper.New(),
		cmtcfg.DefaultConfig(),
		log.NewCustomLogger(zerolog.New(tw)),
	)
	sCtx.Config.SetRoot(homeDir)

	cCtx := (client.Context{}).WithHomeDir(homeDir)

	ctx := context.WithValue(context.Background(), server.ServerContextKey, sCtx)
	ctx = context.WithValue(ctx, client.ClientContextKey, &cCtx)

	return &ExportSystem{
		sys:     sys,
		Ctx:     ctx,
		Sctx:    sCtx,
		Cctx:    cCtx,
		HomeDir: homeDir,
	}
}

// Run wraps (*cmdtest.System).RunC, providing e's context.
func (s *ExportSystem) Run(args ...string) cmdtest.RunResult {
	return s.sys.RunC(s.Ctx, args...)
}

// MustRun wraps (*cmdtest.System).MustRunC, providing e's context.
func (s *ExportSystem) MustRun(t *testing.T, args ...string) cmdtest.RunResult {
	return s.sys.MustRunC(t, s.Ctx, args...)
}

// isZeroExportedApp reports whether all fields of a are unset.
//
// This is for the mockExporter to check if a return value was ever set.
func isZeroExportedApp(a types.ExportedApp) bool {
	return a.AppState == nil &&
		len(a.Validators) == 0 &&
		a.Height == 0 &&
		a.ConsensusParams == cmtproto.ConsensusParams{}
}

// mockExporter provides an Export method matching server/types.AppExporter,
// and it tracks relevant arguments when that method is called.
type mockExporter struct {
	// The values to return from Export().
	ExportApp types.ExportedApp
	Err       error

	// Whether Export was called at all.
	WasCalled bool

	// Called tracks the interesting arguments passed to Export().
	Called struct {
		Height           int64
		ForZeroHeight    bool
		JailAllowedAddrs []string
		ModulesToExport  []string
	}
}

// SetDefaultExportApp sets a valid ExportedApp to be returned
// when e.Export is called.
func (e *mockExporter) SetDefaultExportApp() {
	e.ExportApp = types.ExportedApp{
		ConsensusParams: cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{
				MaxBytes: 5 * 1024 * 1024,
				MaxGas:   -1,
			},
			Evidence: &cmtproto.EvidenceParams{
				MaxAgeNumBlocks: 100,
				MaxAgeDuration:  time.Hour,
				MaxBytes:        1024 * 1024,
			},
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: []string{cmttypes.ABCIPubKeyTypeEd25519},
			},
		},
	}
}

// Export satisfies the server/types.AppExporter function type.
//
// e tracks relevant arguments under the e.Called struct.
//
// Export panics if neither e.ExportApp nor e.Err have been set.
func (e *mockExporter) Export(
	logger log.Logger,
	db dbm.DB,
	traceWriter io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	opts types.AppOptions,
	modulesToExport []string,
) (types.ExportedApp, error) {
	if e.Err == nil && isZeroExportedApp(e.ExportApp) {
		panic(fmt.Errorf("(*mockExporter).Export called without setting e.ExportApp or e.Err"))
	}
	e.WasCalled = true

	e.Called.Height = height
	e.Called.ForZeroHeight = forZeroHeight
	e.Called.JailAllowedAddrs = jailAllowedAddrs
	e.Called.ModulesToExport = modulesToExport

	return e.ExportApp, e.Err
}

func TestExportCLI(t *testing.T) {
	// Use t.Parallel in all of the subtests,
	// because they all read from disk and risk blocking on io.

	t.Run("fail on missing genesis file", func(t *testing.T) {
		t.Parallel()

		e := new(mockExporter)
		sys := NewExportSystem(t, e.Export)

		res := sys.Run("export")
		require.Error(t, res.Err)
		require.Truef(t, os.IsNotExist(res.Err), "expected resulting error to be os.IsNotExist, got %T (%v)", res.Err, res.Err)

		require.False(t, e.WasCalled)
	})

	t.Run("prints to stdout by default", func(t *testing.T) {
		t.Parallel()

		e := new(mockExporter)
		e.SetDefaultExportApp()

		sys := NewExportSystem(t, e.Export)
		_ = sys.MustRun(t, "init", "some_moniker")
		res := sys.MustRun(t, "export")

		require.Empty(t, res.Stderr.String())

		CheckExportedGenesis(t, res.Stdout.Bytes())
	})

	t.Run("passes expected default values to the AppExporter", func(t *testing.T) {
		t.Parallel()

		e := new(mockExporter)
		e.SetDefaultExportApp()

		sys := NewExportSystem(t, e.Export)
		_ = sys.MustRun(t, "init", "some_moniker")
		_ = sys.MustRun(t, "export")

		require.True(t, e.WasCalled)

		require.Equal(t, int64(-1), e.Called.Height)
		require.False(t, e.Called.ForZeroHeight)
		require.Empty(t, e.Called.JailAllowedAddrs)
		require.Empty(t, e.Called.ModulesToExport)
	})

	t.Run("passes flag values to the AppExporter", func(t *testing.T) {
		t.Parallel()

		e := new(mockExporter)
		e.SetDefaultExportApp()

		sys := NewExportSystem(t, e.Export)
		_ = sys.MustRun(t, "init", "some_moniker")
		_ = sys.MustRun(t, "export",
			"--height=100",
			"--jail-allowed-addrs", "addr1,addr2",
			"--modules-to-export", "foo,bar",
		)

		require.True(t, e.WasCalled)

		require.Equal(t, int64(100), e.Called.Height)
		require.False(t, e.Called.ForZeroHeight)
		require.Equal(t, []string{"addr1", "addr2"}, e.Called.JailAllowedAddrs)
		require.Equal(t, []string{"foo", "bar"}, e.Called.ModulesToExport)
	})

	t.Run("passes --for-zero-height to the AppExporter", func(t *testing.T) {
		t.Parallel()

		e := new(mockExporter)
		e.SetDefaultExportApp()

		sys := NewExportSystem(t, e.Export)
		_ = sys.MustRun(t, "init", "some_moniker")
		_ = sys.MustRun(t, "export", "--for-zero-height")

		require.True(t, e.WasCalled)

		require.Equal(t, int64(-1), e.Called.Height)
		require.True(t, e.Called.ForZeroHeight)
		require.Empty(t, e.Called.JailAllowedAddrs)
		require.Empty(t, e.Called.ModulesToExport)
	})

	t.Run("prints to a given file with --output-document", func(t *testing.T) {
		t.Parallel()

		e := new(mockExporter)
		e.SetDefaultExportApp()

		sys := NewExportSystem(t, e.Export)
		_ = sys.MustRun(t, "init", "some_moniker")

		outDir := t.TempDir()
		outFile := filepath.Join(outDir, "export.json")

		res := sys.MustRun(t, "export", "--output-document", outFile)

		require.Empty(t, res.Stderr.String())
		require.Empty(t, res.Stdout.String())

		j, err := os.ReadFile(outFile)
		require.NoError(t, err)

		CheckExportedGenesis(t, j)
	})

	t.Run("prints genesis to stdout when no app exporter defined", func(t *testing.T) {
		t.Parallel()

		sys := NewExportSystem(t, nil)
		_ = sys.MustRun(t, "init", "some_moniker")

		res := sys.MustRun(t, "export")

		require.Contains(t, res.Stderr.String(), "WARNING: App exporter not defined.")

		origGenesis, err := os.ReadFile(filepath.Join(sys.HomeDir, "config", "genesis.json"))
		require.NoError(t, err)

		out := res.Stdout.Bytes()

		require.Equal(t, origGenesis, out)
	})

	t.Run("returns app exporter error", func(t *testing.T) {
		t.Parallel()

		e := new(mockExporter)
		e.Err = fmt.Errorf("whoopsie")

		sys := NewExportSystem(t, e.Export)
		_ = sys.MustRun(t, "init", "some_moniker")

		res := sys.Run("export")

		require.ErrorIs(t, res.Err, e.Err)
	})

	t.Run("rejects positional arguments", func(t *testing.T) {
		t.Parallel()

		e := new(mockExporter)
		e.SetDefaultExportApp()

		sys := NewExportSystem(t, e.Export)
		_ = sys.MustRun(t, "init", "some_moniker")

		outDir := t.TempDir()
		outFile := filepath.Join(outDir, "export.json")

		res := sys.Run("export", outFile)
		require.Error(t, res.Err)

		require.NoFileExists(t, outFile)
	})
}

// CheckExportedGenesis fails t if j cannot be unmarshaled into a valid AppGenesis.
func CheckExportedGenesis(t *testing.T, j []byte) {
	t.Helper()

	var ag genutiltypes.AppGenesis
	require.NoError(t, json.Unmarshal(j, &ag))

	require.NotEmpty(t, ag.AppName)
	require.NotZero(t, ag.GenesisTime)
	require.NotEmpty(t, ag.ChainID)
	require.NotNil(t, ag.Consensus)
}
