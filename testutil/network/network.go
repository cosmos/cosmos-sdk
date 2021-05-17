package network

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmcfg "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/node"
	tmclient "github.com/tendermint/tendermint/rpc/client"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// package-wide network lock to only allow one test network at a time
var lock = new(sync.Mutex)

// AppConstructor defines a function which accepts a network configuration and
// creates an ABCI Application to provide to Tendermint.
type AppConstructor = func(val Validator) servertypes.Application

// NewAppConstructor returns a new simapp AppConstructor
func NewAppConstructor(encodingCfg params.EncodingConfig) AppConstructor {
	return func(val Validator) servertypes.Application {
		return simapp.NewSimApp(
			val.Ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
			encodingCfg,
			simapp.EmptyAppOptions{},
			baseapp.SetPruning(storetypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
}

// Config defines the necessary configuration used to bootstrap and start an
// in-process local testing network.
type Config struct {
	Codec             codec.Marshaler
	LegacyAmino       *codec.LegacyAmino // TODO: Remove!
	InterfaceRegistry codectypes.InterfaceRegistry

	TxConfig         client.TxConfig
	AccountRetriever client.AccountRetriever
	AppConstructor   AppConstructor             // the ABCI application constructor
	GenesisState     map[string]json.RawMessage // custom gensis state to provide
	TimeoutCommit    time.Duration              // the consensus commitment timeout
	ChainID          string                     // the network chain-id
	NumValidators    int                        // the total number of validators to create and bond
	BondDenom        string                     // the staking bond denomination
	MinGasPrices     string                     // the minimum gas prices each validator will accept
	AccountTokens    sdk.Int                    // the amount of unique validator tokens (e.g. 1000node0)
	StakingTokens    sdk.Int                    // the amount of tokens each validator has available to stake
	BondedTokens     sdk.Int                    // the amount of tokens each validator stakes
	PruningStrategy  string                     // the pruning strategy each validator will have
	EnableLogging    bool                       // enable Tendermint logging to STDOUT
	CleanupDir       bool                       // remove base temporary directory during cleanup
	SigningAlgo      string                     // signing algorithm for keys
	KeyringOptions   []keyring.Option
}

// DefaultConfig returns a sane default configuration suitable for nearly all
// testing requirements.
func DefaultConfig() Config {
	encCfg := simapp.MakeTestEncodingConfig()

	return Config{
		Codec:             encCfg.Marshaler,
		TxConfig:          encCfg.TxConfig,
		LegacyAmino:       encCfg.Amino,
		InterfaceRegistry: encCfg.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor:    NewAppConstructor(encCfg),
		GenesisState:      simapp.ModuleBasics.DefaultGenesis(encCfg.Marshaler),
		TimeoutCommit:     2 * time.Second,
		ChainID:           "chain-" + tmrand.NewRand().Str(6),
		NumValidators:     4,
		BondDenom:         sdk.DefaultBondDenom,
		MinGasPrices:      fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
		AccountTokens:     sdk.TokensFromConsensusPower(1000),
		StakingTokens:     sdk.TokensFromConsensusPower(500),
		BondedTokens:      sdk.TokensFromConsensusPower(100),
		PruningStrategy:   storetypes.PruningOptionNothing,
		CleanupDir:        true,
		SigningAlgo:       string(hd.Secp256k1Type),
		KeyringOptions:    []keyring.Option{},
	}
}

type (
	// Network defines a local in-process testing network using SimApp. It can be
	// configured to start any number of validators, each with its own RPC and API
	// clients. Typically, this test network would be used in client and integration
	// testing where user input is expected.
	//
	// Note, due to Tendermint constraints in regards to RPC functionality, there
	// may only be one test network running at a time. Thus, any caller must be
	// sure to Cleanup after testing is finished in order to allow other tests
	// to create networks. In addition, only the first validator will have a valid
	// RPC and API server/client.
	Network struct {
		T          *testing.T
		BaseDir    string
		Validators []*Validator

		Config Config
	}

	// Validator defines an in-process Tendermint validator node. Through this object,
	// a client can make RPC and API calls and interact with any client command
	// or handler.
	Validator struct {
		AppConfig  *srvconfig.Config
		ClientCtx  client.Context
		Ctx        *server.Context
		Dir        string
		NodeID     string
		PubKey     cryptotypes.PubKey
		Moniker    string
		APIAddress string
		RPCAddress string
		P2PAddress string
		Address    sdk.AccAddress
		ValAddress sdk.ValAddress
		RPCClient  tmclient.Client

		tmNode *node.Node
		api    *api.Server
		grpc   *grpc.Server
	}
)

// New creates a new Network for integration tests.
func New(t *testing.T, cfg Config) *Network {
	// only one caller/test can create and use a network at a time
	t.Log("acquiring test network lock")
	lock.Lock()

	baseDir, err := ioutil.TempDir(t.TempDir(), cfg.ChainID)
	require.NoError(t, err)
	t.Logf("created temporary directory: %s", baseDir)

	network := &Network{
		T:          t,
		BaseDir:    baseDir,
		Validators: make([]*Validator, cfg.NumValidators),
		Config:     cfg,
	}

	t.Log("preparing test network...")

	monikers := make([]string, cfg.NumValidators)
	nodeIDs := make([]string, cfg.NumValidators)
	valPubKeys := make([]cryptotypes.PubKey, cfg.NumValidators)

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
		genFiles    []string
	)

	buf := bufio.NewReader(os.Stdin)

	// generate private keys, node IDs, and initial transactions
	for i := 0; i < cfg.NumValidators; i++ {
		appCfg := srvconfig.DefaultConfig()
		appCfg.Pruning = cfg.PruningStrategy
		appCfg.MinGasPrices = cfg.MinGasPrices
		appCfg.API.Enable = true
		appCfg.API.Swagger = false
		appCfg.Telemetry.Enabled = false

		ctx := server.NewDefaultContext()
		tmCfg := ctx.Config
		tmCfg.Consensus.TimeoutCommit = cfg.TimeoutCommit

		// Only allow the first validator to expose an RPC, API and gRPC
		// server/client due to Tendermint in-process constraints.
		apiAddr := ""
		tmCfg.RPC.ListenAddress = ""
		appCfg.GRPC.Enable = false
		if i == 0 {
			apiListenAddr, _, err := server.FreeTCPAddr()
			require.NoError(t, err)
			appCfg.API.Address = apiListenAddr

			apiURL, err := url.Parse(apiListenAddr)
			require.NoError(t, err)
			apiAddr = fmt.Sprintf("http://%s:%s", apiURL.Hostname(), apiURL.Port())

			rpcAddr, _, err := server.FreeTCPAddr()
			require.NoError(t, err)
			tmCfg.RPC.ListenAddress = rpcAddr

			_, grpcPort, err := server.FreeTCPAddr()
			require.NoError(t, err)
			appCfg.GRPC.Address = fmt.Sprintf("0.0.0.0:%s", grpcPort)
			appCfg.GRPC.Enable = true
		}

		logger := log.NewNopLogger()
		if cfg.EnableLogging {
			logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
			logger, _ = tmflags.ParseLogLevel("info", logger, tmcfg.DefaultLogLevel)
		}

		ctx.Logger = logger

		nodeDirName := fmt.Sprintf("node%d", i)
		nodeDir := filepath.Join(network.BaseDir, nodeDirName, "simd")
		clientDir := filepath.Join(network.BaseDir, nodeDirName, "simcli")
		gentxsDir := filepath.Join(network.BaseDir, "gentxs")

		require.NoError(t, os.MkdirAll(filepath.Join(nodeDir, "config"), 0755))
		require.NoError(t, os.MkdirAll(clientDir, 0755))

		tmCfg.SetRoot(nodeDir)
		tmCfg.Moniker = nodeDirName
		monikers[i] = nodeDirName

		proxyAddr, _, err := server.FreeTCPAddr()
		require.NoError(t, err)
		tmCfg.ProxyApp = proxyAddr

		p2pAddr, _, err := server.FreeTCPAddr()
		require.NoError(t, err)
		tmCfg.P2P.ListenAddress = p2pAddr
		tmCfg.P2P.AddrBookStrict = false
		tmCfg.P2P.AllowDuplicateIP = true

		nodeID, pubKey, err := genutil.InitializeNodeValidatorFiles(tmCfg)
		require.NoError(t, err)
		nodeIDs[i] = nodeID
		valPubKeys[i] = pubKey

		kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, clientDir, buf, cfg.KeyringOptions...)
		require.NoError(t, err)

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(cfg.SigningAlgo, keyringAlgos)
		require.NoError(t, err)

		addr, secret, err := server.GenerateSaveCoinKey(kb, nodeDirName, true, algo)
		require.NoError(t, err)

		info := map[string]string{"secret": secret}
		infoBz, err := json.Marshal(info)
		require.NoError(t, err)

		// save private key seed words
		require.NoError(t, writeFile(fmt.Sprintf("%v.json", "key_seed"), clientDir, infoBz))

		balances := sdk.NewCoins(
			sdk.NewCoin(fmt.Sprintf("%stoken", nodeDirName), cfg.AccountTokens),
			sdk.NewCoin(cfg.BondDenom, cfg.StakingTokens),
		)

		genFiles = append(genFiles, tmCfg.GenesisFile())
		genBalances = append(genBalances, banktypes.Balance{Address: addr.String(), Coins: balances.Sort()})
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(addr, nil, 0, 0))

		commission, err := sdk.NewDecFromStr("0.5")
		require.NoError(t, err)

		createValMsg, err := stakingtypes.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			valPubKeys[i],
			sdk.NewCoin(cfg.BondDenom, cfg.BondedTokens),
			stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
			stakingtypes.NewCommissionRates(commission, sdk.OneDec(), sdk.OneDec()),
			sdk.OneInt(),
		)
		require.NoError(t, err)

		p2pURL, err := url.Parse(p2pAddr)
		require.NoError(t, err)

		memo := fmt.Sprintf("%s@%s:%s", nodeIDs[i], p2pURL.Hostname(), p2pURL.Port())
		fee := sdk.NewCoins(sdk.NewCoin(fmt.Sprintf("%stoken", nodeDirName), sdk.NewInt(0)))
		txBuilder := cfg.TxConfig.NewTxBuilder()
		require.NoError(t, txBuilder.SetMsgs(createValMsg))
		txBuilder.SetFeeAmount(fee)    // Arbitrary fee
		txBuilder.SetGasLimit(1000000) // Need at least 100386
		txBuilder.SetMemo(memo)

		txFactory := tx.Factory{}
		txFactory = txFactory.
			WithChainID(cfg.ChainID).
			WithMemo(memo).
			WithKeybase(kb).
			WithTxConfig(cfg.TxConfig)

		err = tx.Sign(txFactory, nodeDirName, txBuilder, true)
		require.NoError(t, err)

		txBz, err := cfg.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		require.NoError(t, err)
		require.NoError(t, writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz))

		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), appCfg)

		clientCtx := client.Context{}.
			WithKeyringDir(clientDir).
			WithKeyring(kb).
			WithHomeDir(tmCfg.RootDir).
			WithChainID(cfg.ChainID).
			WithInterfaceRegistry(cfg.InterfaceRegistry).
			WithJSONMarshaler(cfg.Codec).
			WithLegacyAmino(cfg.LegacyAmino).
			WithTxConfig(cfg.TxConfig).
			WithAccountRetriever(cfg.AccountRetriever)

		network.Validators[i] = &Validator{
			AppConfig:  appCfg,
			ClientCtx:  clientCtx,
			Ctx:        ctx,
			Dir:        filepath.Join(network.BaseDir, nodeDirName),
			NodeID:     nodeID,
			PubKey:     pubKey,
			Moniker:    nodeDirName,
			RPCAddress: tmCfg.RPC.ListenAddress,
			P2PAddress: tmCfg.P2P.ListenAddress,
			APIAddress: apiAddr,
			Address:    addr,
			ValAddress: sdk.ValAddress(addr),
		}
	}

	require.NoError(t, initGenFiles(cfg, genAccounts, genBalances, genFiles))
	require.NoError(t, collectGenFiles(cfg, network.Validators, network.BaseDir))

	t.Log("starting test network...")
	for _, v := range network.Validators {
		require.NoError(t, startInProcess(cfg, v))
	}

	t.Log("started test network")

	// Ensure we cleanup incase any test was abruptly halted (e.g. SIGINT) as any
	// defer in a test would not be called.
	server.TrapSignal(network.Cleanup)

	return network
}

// LatestHeight returns the latest height of the network or an error if the
// query fails or no validators exist.
func (n *Network) LatestHeight() (int64, error) {
	if len(n.Validators) == 0 {
		return 0, errors.New("no validators available")
	}

	status, err := n.Validators[0].RPCClient.Status(context.Background())
	if err != nil {
		return 0, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

// WaitForHeight performs a blocking check where it waits for a block to be
// committed after a given block. If that height is not reached within a timeout,
// an error is returned. Regardless, the latest height queried is returned.
func (n *Network) WaitForHeight(h int64) (int64, error) {
	return n.WaitForHeightWithTimeout(h, 10*time.Second)
}

// WaitForHeightWithTimeout is the same as WaitForHeight except the caller can
// provide a custom timeout.
func (n *Network) WaitForHeightWithTimeout(h int64, t time.Duration) (int64, error) {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(t)

	if len(n.Validators) == 0 {
		return 0, errors.New("no validators available")
	}

	var latestHeight int64
	val := n.Validators[0]

	for {
		select {
		case <-timeout:
			ticker.Stop()
			return latestHeight, errors.New("timeout exceeded waiting for block")
		case <-ticker.C:
			status, err := val.RPCClient.Status(context.Background())
			if err == nil && status != nil {
				latestHeight = status.SyncInfo.LatestBlockHeight
				if latestHeight >= h {
					return latestHeight, nil
				}
			}
		}
	}
}

// WaitForNextBlock waits for the next block to be committed, returning an error
// upon failure.
func (n *Network) WaitForNextBlock() error {
	lastBlock, err := n.LatestHeight()
	if err != nil {
		return err
	}

	_, err = n.WaitForHeight(lastBlock + 1)
	if err != nil {
		return err
	}

	return err
}

// Cleanup removes the root testing (temporary) directory and stops both the
// Tendermint and API services. It allows other callers to create and start
// test networks. This method must be called when a test is finished, typically
// in a defer.
func (n *Network) Cleanup() {
	defer func() {
		lock.Unlock()
		n.T.Log("released test network lock")
	}()

	n.T.Log("cleaning up test network...")

	for _, v := range n.Validators {
		if v.tmNode != nil && v.tmNode.IsRunning() {
			_ = v.tmNode.Stop()
		}

		if v.api != nil {
			_ = v.api.Close()
		}

		if v.grpc != nil {
			v.grpc.Stop()
		}
	}

	if n.Config.CleanupDir {
		_ = os.RemoveAll(n.BaseDir)
	}

	n.T.Log("finished cleaning up test network")
}
