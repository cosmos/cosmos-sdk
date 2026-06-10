package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/rpc/client/local"
	cmtrpcserver "github.com/cometbft/cometbft/rpc/jsonrpc/server"
	cmttypes "github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/gorilla/handlers"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/net/netutil"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	servercmtlog "github.com/cosmos/cosmos-sdk/server/log"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func startInProcess(cfg Config, val *Validator) error {
	logger := val.Ctx.Logger
	cmtCfg := val.Ctx.Config
	cmtCfg.Instrumentation.Prometheus = false

	if err := val.AppConfig.ValidateBasic(); err != nil {
		return err
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(cmtCfg.NodeKeyFile())
	if err != nil {
		return err
	}

	app := cfg.AppConstructor(*val)
	val.app = app

	appGenesisProvider := func() (*cmttypes.GenesisDoc, error) {
		appGenesis, err := genutiltypes.AppGenesisFromFile(cmtCfg.GenesisFile())
		if err != nil {
			return nil, err
		}

		return appGenesis.ToGenesisDoc()
	}

	cmtApp := server.NewCometABCIWrapper(app)
	tmNode, err := node.NewNode( //resleak:notresource
		cmtCfg,
		pvm.LoadOrGenFilePV(cmtCfg.PrivValidatorKeyFile(), cmtCfg.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(cmtApp),
		appGenesisProvider,
		cmtcfg.DefaultDBProvider,
		node.DefaultMetricsProvider(cmtCfg.Instrumentation),
		servercmtlog.CometLoggerWrapper{Logger: logger.With("module", val.Moniker)},
	)
	if err != nil {
		return err
	}

	if err := tmNode.Start(); err != nil {
		return err
	}
	val.tmNode = tmNode

	// Update P2PAddress with the actual port CometBFT bound to.
	if ni, ok := tmNode.NodeInfo().(p2p.DefaultNodeInfo); ok && ni.ListenAddr != "" {
		val.P2PAddress = "tcp://" + ni.ListenAddr
	}

	if val.RPCAddress != "" {
		val.RPCClient = local.New(tmNode)
	}

	// We'll need a RPC client if the validator exposes a gRPC or REST endpoint.
	if val.APIAddress != "" || val.AppConfig.GRPC.Enable {
		val.ClientCtx = val.ClientCtx.
			WithClient(val.RPCClient)

		app.RegisterTxService(val.ClientCtx)
		app.RegisterTendermintService(val.ClientCtx)
		app.RegisterNodeService(val.ClientCtx, *val.AppConfig)
	}

	ctx := context.Background()
	ctx, val.cancelFn = context.WithCancel(ctx)
	val.errGroup, ctx = errgroup.WithContext(ctx)

	grpcCfg := val.AppConfig.GRPC

	if grpcCfg.Enable {
		grpcLogger := logger.With(log.ModuleKey, "grpc-server")

		// Bind an ephemeral port; update the address fields; serve on the same fd.
		var grpcLis net.Listener
		if val.AppConfig.GRPC.Address == "" {
			grpcLis, err = net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return fmt.Errorf("failed to bind gRPC listener: %w", err)
			}
			val.AppConfig.GRPC.Address = grpcLis.Addr().String()
			grpcCfg.Address = grpcLis.Addr().String()
		}

		var grpcSrv *grpc.Server
		grpcSrv, val.ClientCtx, err = servergrpc.NewGRPCServerAndContext(val.ClientCtx, app, grpcCfg, grpcLogger)
		if err != nil {
			if grpcLis != nil {
				grpcLis.Close()
			}
			return err
		}

		val.errGroup.Go(func() error {
			if grpcLis != nil {
				return serveGRPC(ctx, grpcLogger, grpcSrv, grpcLis)
			}
			return servergrpc.StartGRPCServer(ctx, grpcLogger, grpcCfg, grpcSrv)
		})

		val.grpc = grpcSrv
	}

	if val.AppConfig.API.Enable {
		apiLogger := logger.With(log.ModuleKey, "api-server")

		var apiLis net.Listener
		if val.AppConfig.API.Address == "" {
			apiLis, err = net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return fmt.Errorf("failed to bind API listener: %w", err)
			}
			val.AppConfig.API.Address = "tcp://" + apiLis.Addr().String()
			val.APIAddress = "http://" + apiLis.Addr().String()
		}

		apiSrv := api.New(val.ClientCtx, apiLogger, val.grpc)
		app.RegisterAPIRoutes(apiSrv, val.AppConfig.API)

		val.errGroup.Go(func() error {
			if apiLis != nil {
				return serveAPI(ctx, apiSrv, apiLogger, *val.AppConfig, apiLis)
			}
			return apiSrv.Start(ctx, *val.AppConfig)
		})

		val.api = apiSrv
	}

	return nil
}

func collectGenFiles(cfg Config, vals []*Validator, outputDir string) error {
	genTime := cmttime.Now()

	for i := range cfg.NumValidators {
		cmtCfg := vals[i].Ctx.Config

		nodeDir := filepath.Join(outputDir, vals[i].Moniker, "simd")
		gentxsDir := filepath.Join(outputDir, "gentxs")

		cmtCfg.Moniker = vals[i].Moniker
		cmtCfg.SetRoot(nodeDir)

		initCfg := genutiltypes.NewInitConfig(cfg.ChainID, gentxsDir, vals[i].NodeID, vals[i].PubKey)

		genFile := cmtCfg.GenesisFile()
		appGenesis, err := genutiltypes.AppGenesisFromFile(genFile)
		if err != nil {
			return err
		}

		appState, err := genutil.GenAppStateFromConfig(cfg.Codec, cfg.TxConfig,
			cmtCfg, initCfg, appGenesis, banktypes.GenesisBalancesIterator{}, genutiltypes.DefaultMessageValidator, cfg.TxConfig.SigningContext().ValidatorAddressCodec())
		if err != nil {
			return err
		}

		// overwrite each validator's genesis file to have a canonical genesis
		// time, preserving any custom ConsensusParams (e.g. the
		// Validator.PubKeyTypes the network was bootstrapped with).
		// genutil.ExportGenesisFileWithTime is unsuitable here because it
		// builds a fresh AppGenesis whose ConsensusParams default back to
		// `[ed25519]`, wiping ml_dsa_65 or any other opt-in key type.
		appGenesis.GenesisTime = genTime
		appGenesis.AppState = appState
		if appGenesis.Consensus != nil {
			appGenesis.Consensus.Validators = nil
		}
		if err := genutil.ExportGenesisFile(appGenesis, genFile); err != nil {
			return err
		}
	}

	return nil
}

func initGenFiles(cfg Config, genAccounts []authtypes.GenesisAccount, genBalances []banktypes.Balance, genFiles []string) error {
	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = append(authGenState.Accounts, accounts...)
	cfg.GenesisState[authtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&authGenState)

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[banktypes.ModuleName], &bankGenState)

	bankGenState.Balances = append(bankGenState.Balances, genBalances...)
	cfg.GenesisState[banktypes.ModuleName] = cfg.Codec.MustMarshalJSON(&bankGenState)

	appGenStateJSON, err := json.MarshalIndent(cfg.GenesisState, "", "  ")
	if err != nil {
		return err
	}

	consensus := &genutiltypes.ConsensusGenesis{
		Validators: nil,
	}
	// When the caller pins validators to a non-default consensus key type
	// (e.g. ml_dsa_65), pre-populate ConsensusParams so the staking module
	// won't reject MsgCreateValidator at InitChain with
	// "validator pubkey type is not supported".
	if cfg.ValidatorConsensusKeyType != "" {
		params := cmttypes.DefaultConsensusParams()
		params.Validator.PubKeyTypes = []string{cfg.ValidatorConsensusKeyType}
		consensus.Params = params
	}
	appGenesis := genutiltypes.AppGenesis{
		ChainID:   cfg.ChainID,
		AppState:  appGenStateJSON,
		Consensus: consensus,
	}

	// generate empty genesis files for each validator and save
	for i := range cfg.NumValidators {
		if err := appGenesis.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}

	return nil
}

func writeFile(name, dir string, contents []byte) error {
	file := filepath.Join(dir, name)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("could not create directory %q: %w", dir, err)
	}

	if err := os.WriteFile(file, contents, 0o600); err != nil {
		return err
	}

	return nil
}

// serveGRPC serves grpcSrv on lis with context-driven shutdown.
func serveGRPC(ctx context.Context, logger log.Logger, grpcSrv *grpc.Server, lis net.Listener) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting gRPC server...", "address", lis.Addr())
		errCh <- grpcSrv.Serve(lis)
	}()
	select {
	case <-ctx.Done():
		logger.Info("stopping gRPC server...", "address", lis.Addr())
		grpcSrv.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

// serveAPI serves an already-configured api.Server on lis.
func serveAPI(ctx context.Context, srv *api.Server, logger log.Logger, cfg srvconfig.Config, lis net.Listener) error {
	cmtCfg := cmtrpcserver.DefaultConfig()
	cmtCfg.MaxOpenConnections = int(cfg.API.MaxOpenConnections)
	cmtCfg.ReadTimeout = time.Duration(cfg.API.RPCReadTimeout) * time.Second
	cmtCfg.WriteTimeout = time.Duration(cfg.API.RPCWriteTimeout) * time.Second
	cmtCfg.MaxBodyBytes = int64(cfg.API.RPCMaxBodyBytes)

	if cmtCfg.MaxOpenConnections > 0 {
		lis = netutil.LimitListener(lis, cmtCfg.MaxOpenConnections)
	}

	if cfg.GRPC.Enable && cfg.GRPCWeb.Enable {
		var options []grpcweb.Option
		if cfg.API.EnableUnsafeCORS {
			options = append(options, grpcweb.WithOriginFunc(func(string) bool { return true }))
		}
		wrappedGrpc := grpcweb.WrapServer(srv.GRPCSrv, options...)
		srv.Router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if wrappedGrpc.IsGrpcWebRequest(req) {
				wrappedGrpc.ServeHTTP(w, req)
				return
			}
			srv.GRPCGatewayRouter.ServeHTTP(w, req)
		}))
	}
	srv.Router.PathPrefix("/").Handler(srv.GRPCGatewayRouter)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting API server...", "address", lis.Addr())
		if cfg.API.EnableUnsafeCORS {
			allowAllCORS := handlers.CORS(handlers.AllowedHeaders([]string{"Content-Type"}))
			errCh <- cmtrpcserver.Serve(lis, allowAllCORS(srv.Router), servercmtlog.CometLoggerWrapper{Logger: logger}, cmtCfg)
		} else {
			errCh <- cmtrpcserver.Serve(lis, srv.Router, servercmtlog.CometLoggerWrapper{Logger: logger}, cmtCfg)
		}
	}()
	select {
	case <-ctx.Done():
		logger.Info("stopping API server...", "address", lis.Addr())
		return lis.Close()
	case err := <-errCh:
		return err
	}
}
