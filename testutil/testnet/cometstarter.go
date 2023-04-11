package testnet

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"cosmossdk.io/log"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	cmttypes "github.com/cometbft/cometbft/types"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// CometStarter offers a builder-pattern interface to
// starting a Comet instance with an ABCI application running alongside.
//
// As CometStart is more broadly used in the codebase,
// the number of available methods on CometStarter will grow.
type CometStarter struct {
	logger log.Logger

	app abcitypes.Application

	cfg        *cmtcfg.Config
	valPrivKey cmted25519.PrivKey
	genesis    []byte

	rootDir string

	rpcListen bool

	tcpAddrChooser func() string

	startTries int
}

// NewCometStarter accepts a minimal set of arguments to start comet with an ABCI app.
// For further configuration, chain other CometStarter methods before calling Start:
//
//     NewCometStarter(...).Logger(...).Start()
func NewCometStarter(
	app abcitypes.Application,
	cfg *cmtcfg.Config,
	valPrivKey cmted25519.PrivKey,
	genesis []byte,
	rootDir string,
) *CometStarter {
	cfg.SetRoot(rootDir)

	// CometStarter won't work without these settings,
	// so set them unconditionally.
	cfg.P2P.AllowDuplicateIP = true
	cfg.P2P.AddrBookStrict = false

	// For now, we disallow RPC listening.
	// Comet v0.37 uses a global value such that multiple comet nodes in one process
	// end up contending over one "rpc environment" and only the last-started validator
	// will control the RPC service.
	//
	// The "rpc environment" was removed as a global in
	// https://github.com/cometbft/cometbft/commit/3324f49fb7e7b40189726746493e83b82a61b558
	// which is due to land in v0.38.
	//
	// At that point, we should keep the default as RPC off,
	// but we should add a RPCListen method to opt in to enabling it.

	// If RPC.ListenAddress is the default value, clear it.
	const defaultRPCListenAddr = "tcp://127.0.0.1:26657"
	if cfg.RPC.ListenAddress == defaultRPCListenAddr {
		cfg.RPC.ListenAddress = ""
	}

	// Then if it was set to anything other than empty or the default value,
	// fail with a clear explanation on how to enable RPC.
	// The RPCListen method must be used in order to correctly pick an available listen address.
	if cfg.RPC.ListenAddress != "" {
		panic(fmt.Errorf("NewCometStarter: cfg.RPC.ListenAddress must be empty (but was %q); use (*CometStarter).RPCListen() instead", cfg.RPC.ListenAddress))
	}

	// defaultStartTries is somewhat arbitrary.
	// Occasionally TestCometStarter_PortContention would fail with 10 tries,
	// and bumping it up to 12 makes it almost never fail.
	const defaultStartTries = 12
	return &CometStarter{
		logger: log.NewNopLogger(),

		app: app,

		cfg:        cfg,
		genesis:    genesis,
		valPrivKey: valPrivKey,

		rootDir: rootDir,

		startTries: defaultStartTries,
	}
}

// Logger sets the logger for s and for the eventual started comet instance.
func (s *CometStarter) Logger(logger log.Logger) *CometStarter {
	s.logger = logger
	return s
}

// RPCListen enables the RPC listener service on the underlying Comet node.
// The RPC service must be enabled this way so that s can choose a dynamic port,
// retrying if necessary.
//
// Note that there is a limitation in CometBFT v0.37 that
// prevents more than one RPC server running at a time.
// Once the Cosmos SDK has adopted CometBFT v0.38 or newer,
// that limitation will be removed.
func (s *CometStarter) RPCListen() *CometStarter {
	s.rpcListen = true
	return s
}

// Start returns a started Comet node.
func (s *CometStarter) Start() (n *node.Node, err error) {
	if s.rpcListen {
		if err := globalCometMu.Acquire(); err != nil {
			return nil, err
		}

		// Wrap this defer in an anonymous function so we don't immediately evaluate n,
		// which would always be nil at thi spoint.
		defer func() {
			globalCometMu.Release(n)
		}()
	}

	fpv, nodeKey, err := s.initDisk()
	if err != nil {
		return nil, err
	}

	appGenesisProvider := func() (*cmttypes.GenesisDoc, error) {
		appGenesis, err := genutiltypes.AppGenesisFromFile(s.cfg.GenesisFile())
		if err != nil {
			return nil, err
		}

		return appGenesis.ToGenesisDoc()
	}

	for i := 0; i < s.startTries; i++ {
		s.cfg.P2P.ListenAddress = s.likelyAvailableAddress()
		if s.rpcListen {
			s.cfg.RPC.ListenAddress = s.likelyAvailableAddress()
		}

		n, err := node.NewNode(
			s.cfg,
			fpv,
			nodeKey,
			proxy.NewLocalClientCreator(s.app),
			appGenesisProvider,
			node.DefaultDBProvider,
			node.DefaultMetricsProvider(s.cfg.Instrumentation),
			servercmtlog.CometZeroLogWrapper{Logger: s.logger},
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create comet node: %w", err)
		}

		err = n.Start()
		if err == nil {
			return n, nil
		}

		// Error isn't nil -- if it is EADDRINUSE then we can try again.
		if errors.Is(err, syscall.EADDRINUSE) {
			continue
		}

		// Non-nil error that isn't EADDRINUSE, just return the error.
		return nil, err
	}

	// If we didn't return a node from inside the loop,
	// then we must have exhausted our try limit.
	return nil, fmt.Errorf("failed to start a comet node within %d tries", s.startTries)
}

// initDisk creates the config and data directories on disk,
// and other required files, so that comet and the validator work correctly.
// It also generates a node key for validators.
func (s *CometStarter) initDisk() (cmttypes.PrivValidator, *p2p.NodeKey, error) {
	if err := os.MkdirAll(filepath.Join(s.rootDir, "config"), 0o750); err != nil {
		return nil, nil, fmt.Errorf("failed to make config directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(s.rootDir, "data"), 0o750); err != nil {
		return nil, nil, fmt.Errorf("failed to make data directory: %w", err)
	}

	fpv := privval.NewFilePV(s.valPrivKey, s.cfg.PrivValidatorKeyFile(), s.cfg.PrivValidatorStateFile())
	fpv.Save()

	if err := os.WriteFile(s.cfg.GenesisFile(), s.genesis, 0600); err != nil {
		return nil, nil, fmt.Errorf("failed to write genesis file: %w", err)
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(s.cfg.NodeKeyFile())
	if err != nil {
		return nil, nil, err
	}

	return fpv, nodeKey, nil
}

// TCPAddrChooser sets the function to use when selecting a (likely to be free)
// TCP address for comet's P2P port.
//
// This should only be used when testing CometStarter.
//
// It must return a string in format "tcp://IP:PORT".
func (s *CometStarter) TCPAddrChooser(fn func() string) *CometStarter {
	s.tcpAddrChooser = fn
	return s
}

// likelyAvailableAddress provides a TCP address that is likely to be available
// for comet or other processes to listen on.
//
// Generally, it is better to directly provide a net.Listener that is already bound to an address,
// but unfortunately comet does not offer that as part of its API.
// Instead, we locally bind to :0 and then report that as a "likely available" port.
// If another process steals that port before our comet instance can bind to it,
// the Start method handles retries.
func (s *CometStarter) likelyAvailableAddress() string {
	// If s.TCPAddrChooser was called, use that implementation.
	if s.tcpAddrChooser != nil {
		return s.tcpAddrChooser()
	}

	// Fall back to attempting a random port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Errorf("failed to bind to random port: %w", err))
	}

	defer ln.Close()
	return "tcp://" + ln.Addr().String()
}
