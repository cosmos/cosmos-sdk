package lcdtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	gapp "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	txbuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	bankrest "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrrest "github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	mintrest "github.com/cosmos/cosmos-sdk/x/mint/client/rest"
	paramsrest "github.com/cosmos/cosmos-sdk/x/params/client/rest"
	slashingrest "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingrest "github.com/cosmos/cosmos-sdk/x/staking/client/rest"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	tmrpc "github.com/tendermint/tendermint/rpc/lib/server"
	tmtypes "github.com/tendermint/tendermint/types"
)

// TODO: Make InitializeTestLCD safe to call in multiple tests at the same time
// InitializeTestLCD starts Tendermint and the LCD in process, listening on
// their respective sockets where nValidators is the total number of validators
// and initAddrs are the accounts to initialize with some stake tokens. It
// returns a cleanup function, a set of validator public keys, and a port.
func InitializeLCD(nValidators int, initAddrs []sdk.AccAddress, minting bool, portExt ...string) (
	cleanup func(), valConsPubKeys []crypto.PubKey, valOperAddrs []sdk.ValAddress, port string) {

	if nValidators < 1 {
		panic("InitializeTestLCD must use at least one validator")
	}

	config := GetConfig()
	config.Consensus.TimeoutCommit = 100
	config.Consensus.SkipTimeoutCommit = false
	config.TxIndex.IndexAllTags = true

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger = log.NewFilter(logger, log.AllowError())

	privVal := pvm.LoadOrGenFilePV(config.PrivValidatorKeyFile(),
		config.PrivValidatorStateFile())
	privVal.Reset()

	db := dbm.NewMemDB()
	app := gapp.NewGaiaApp(logger, db, nil, true, 0)
	cdc = gapp.MakeCodec()

	genesisFile := config.GenesisFile()
	genDoc, err := tmtypes.GenesisDocFromFile(genesisFile)
	if err != nil {
		panic(err)
	}
	genDoc.Validators = nil
	err = genDoc.SaveAs(genesisFile)
	if err != nil {
		panic(err)
	}
	genTxs := []json.RawMessage{}

	// append any additional (non-proposing) validators
	var accs []gapp.GenesisAccount
	for i := 0; i < nValidators; i++ {
		operPrivKey := secp256k1.GenPrivKey()
		operAddr := operPrivKey.PubKey().Address()
		pubKey := privVal.GetPubKey()

		power := int64(100)
		if i > 0 {
			pubKey = ed25519.GenPrivKey().PubKey()
			power = 1
		}
		startTokens := sdk.TokensFromTendermintPower(power)

		msg := staking.NewMsgCreateValidator(
			sdk.ValAddress(operAddr),
			pubKey,
			sdk.NewCoin(sdk.DefaultBondDenom, startTokens),
			staking.NewDescription(fmt.Sprintf("validator-%d", i+1), "", "", ""),
			staking.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			sdk.OneInt(),
		)
		stdSignMsg := txbuilder.StdSignMsg{
			ChainID: genDoc.ChainID,
			Msgs:    []sdk.Msg{msg},
		}
		sig, err := operPrivKey.Sign(stdSignMsg.Bytes())
		if err != nil {
			panic(err)
		}
		tx := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{{Signature: sig, PubKey: operPrivKey.PubKey()}}, "")
		txBytes, err := cdc.MarshalJSON(tx)
		if err != nil {
			panic(err)
		}

		genTxs = append(genTxs, txBytes)
		valConsPubKeys = append(valConsPubKeys, pubKey)
		valOperAddrs = append(valOperAddrs, sdk.ValAddress(operAddr))

		accAuth := auth.NewBaseAccountWithAddress(sdk.AccAddress(operAddr))
		accTokens := sdk.TokensFromTendermintPower(150)
		accAuth.Coins = sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, accTokens)}
		accs = append(accs, gapp.NewGenesisAccount(&accAuth))
	}

	appGenState := gapp.NewDefaultGenesisState()
	appGenState.Accounts = accs
	genDoc.AppState, err = cdc.MarshalJSON(appGenState)
	if err != nil {
		panic(err)
	}
	genesisState, err := gapp.GaiaAppGenState(cdc, *genDoc, genTxs)
	if err != nil {
		panic(err)
	}

	// add some tokens to init accounts
	for _, addr := range initAddrs {
		accAuth := auth.NewBaseAccountWithAddress(addr)
		accTokens := sdk.TokensFromTendermintPower(100)
		accAuth.Coins = sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, accTokens)}
		acc := gapp.NewGenesisAccount(&accAuth)
		genesisState.Accounts = append(genesisState.Accounts, acc)
		genesisState.StakingData.Pool.NotBondedTokens = genesisState.StakingData.Pool.NotBondedTokens.Add(accTokens)
	}

	inflationMin := sdk.ZeroDec()
	if minting {
		inflationMin = sdk.MustNewDecFromStr("10000.0")
		genesisState.MintData.Params.InflationMax = sdk.MustNewDecFromStr("15000.0")
	} else {
		genesisState.MintData.Params.InflationMax = inflationMin
	}
	genesisState.MintData.Minter.Inflation = inflationMin
	genesisState.MintData.Params.InflationMin = inflationMin

	// initialize crisis data
	genesisState.CrisisData.ConstantFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)

	// double check inflation is set according to the minting boolean flag
	if minting {
		if !(genesisState.MintData.Params.InflationMax.Equal(sdk.MustNewDecFromStr("15000.0")) &&
			genesisState.MintData.Minter.Inflation.Equal(sdk.MustNewDecFromStr("10000.0")) &&
			genesisState.MintData.Params.InflationMin.Equal(sdk.MustNewDecFromStr("10000.0"))) {
			panic("Mint parameters does not correspond to their defaults")
		}
	} else {
		if !(genesisState.MintData.Params.InflationMax.Equal(sdk.ZeroDec()) &&
			genesisState.MintData.Minter.Inflation.Equal(sdk.ZeroDec()) &&
			genesisState.MintData.Params.InflationMin.Equal(sdk.ZeroDec())) {
			panic("Mint parameters not equal to decimal 0")
		}
	}

	appState, err := codec.MarshalJSONIndent(cdc, genesisState)
	if err != nil {
		panic(err)
	}
	genDoc.AppState = appState

	var listenAddr string

	if len(portExt) == 0 {
		listenAddr, port, err = server.FreeTCPAddr()
		if err != nil {
			panic(err)
		}
	} else {
		listenAddr = fmt.Sprintf("tcp://0.0.0.0:%s", portExt[0])
		port = portExt[0]
	}

	// XXX: Need to set this so LCD knows the tendermint node address!
	viper.Set(client.FlagNode, config.RPC.ListenAddress)
	viper.Set(client.FlagChainID, genDoc.ChainID)
	// TODO Set to false once the upstream Tendermint proof verification issue is fixed.
	viper.Set(client.FlagTrustNode, true)

	node, err := startTM(config, logger, genDoc, privVal, app)
	if err != nil {
		panic(err)
	}

	tests.WaitForNextHeightTM(tests.ExtractPortFromAddress(config.RPC.ListenAddress))
	lcdInstance, err := startLCD(logger, listenAddr, cdc)
	if err != nil {
		panic(err)
	}

	tests.WaitForLCDStart(port)
	tests.WaitForHeight(1, port)

	cleanup = func() {
		logger.Debug("cleaning up LCD initialization")
		err = node.Stop()
		if err != nil {
			logger.Error(err.Error())
		}

		node.Wait()
		err = lcdInstance.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}

	return cleanup, valConsPubKeys, valOperAddrs, port
}

// startTM creates and starts an in-process Tendermint node with memDB and
// in-process ABCI application. It returns the new node or any error that
// occurred.
//
// TODO: Clean up the WAL dir or enable it to be not persistent!
func startTM(
	tmcfg *tmcfg.Config, logger log.Logger, genDoc *tmtypes.GenesisDoc,
	privVal tmtypes.PrivValidator, app abci.Application,
) (*nm.Node, error) {

	genDocProvider := func() (*tmtypes.GenesisDoc, error) { return genDoc, nil }
	dbProvider := func(*nm.DBContext) (dbm.DB, error) { return dbm.NewMemDB(), nil }
	nodeKey, err := p2p.LoadOrGenNodeKey(tmcfg.NodeKeyFile())
	if err != nil {
		return nil, err
	}
	node, err := nm.NewNode(
		tmcfg,
		privVal,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		genDocProvider,
		dbProvider,
		nm.DefaultMetricsProvider(tmcfg.Instrumentation),
		logger.With("module", "node"),
	)
	if err != nil {
		return nil, err
	}

	err = node.Start()
	if err != nil {
		return nil, err
	}

	tests.WaitForRPC(tmcfg.RPC.ListenAddress)
	logger.Info("Tendermint running!")

	return node, err
}

// startLCD starts the LCD.
func startLCD(logger log.Logger, listenAddr string, cdc *codec.Codec) (net.Listener, error) {
	rs := lcd.NewRestServer(cdc)
	registerRoutes(rs)
	listener, err := tmrpc.Listen(listenAddr, tmrpc.DefaultConfig())
	if err != nil {
		return nil, err
	}
	go tmrpc.StartHTTPServer(listener, rs.Mux, logger, tmrpc.DefaultConfig()) //nolint:errcheck
	return listener, nil
}

// NOTE: If making updates here also update cmd/gaia/cmd/gaiacli/main.go
func registerRoutes(rs *lcd.RestServer) {
	rpc.RegisterRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	authrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, auth.StoreKey)
	bankrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	distrrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, distr.StoreKey)
	stakingrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	slashingrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	govrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, paramsrest.ProposalRESTHandler(rs.CliCtx, rs.Cdc))
	mintrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
}

var cdc = amino.NewCodec()

func init() {
	ctypes.RegisterAmino(cdc)
}

// CreateAddr adds an address to the key store and returns an address and seed.
// It also requires that the key could be created.
func CreateAddr(name, password string, kb crkeys.Keybase) (sdk.AccAddress, string, error) {
	var (
		err  error
		info crkeys.Info
		seed string
	)
	info, seed, err = kb.CreateMnemonic(name, crkeys.English, password, crkeys.Secp256k1)
	return sdk.AccAddress(info.GetPubKey().Address()), seed, err
}

// CreateAddr adds multiple address to the key store and returns the addresses and associated seeds in lexographical order by address.
// It also requires that the keys could be created.
func CreateAddrs(kb crkeys.Keybase, numAddrs int) (addrs []sdk.AccAddress, seeds, names, passwords []string) {
	var (
		err  error
		info crkeys.Info
		seed string
	)

	addrSeeds := AddrSeedSlice{}

	for i := 0; i < numAddrs; i++ {
		name := fmt.Sprintf("test%d", i)
		password := "1234567890"
		info, seed, err = kb.CreateMnemonic(name, crkeys.English, password, crkeys.Secp256k1)
		if err != nil {
			panic(err)
		}
		addrSeeds = append(addrSeeds, AddrSeed{Address: sdk.AccAddress(info.GetPubKey().Address()), Seed: seed, Name: name, Password: password})
	}

	sort.Sort(addrSeeds)

	for i := range addrSeeds {
		addrs = append(addrs, addrSeeds[i].Address)
		seeds = append(seeds, addrSeeds[i].Seed)
		names = append(names, addrSeeds[i].Name)
		passwords = append(passwords, addrSeeds[i].Password)
	}

	return addrs, seeds, names, passwords
}

// AddrSeed combines an Address with the mnemonic of the private key to that address
type AddrSeed struct {
	Address  sdk.AccAddress
	Seed     string
	Name     string
	Password string
}

// AddrSeedSlice implements `Interface` in sort package.
type AddrSeedSlice []AddrSeed

func (b AddrSeedSlice) Len() int {
	return len(b)
}

// Less sorts lexicographically by Address
func (b AddrSeedSlice) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i].Address.Bytes(), b[j].Address.Bytes()) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
	}
}

func (b AddrSeedSlice) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// InitClientHome initialises client home dir.
func InitClientHome(dir string) string {
	var err error
	if dir == "" {
		dir, err = ioutil.TempDir("", "lcd_test")
		if err != nil {
			panic(err)
		}
	}
	// TODO: this should be set in NewRestServer
	// and pass down the CLIContext to achieve
	// parallelism.
	viper.Set(cli.HomeFlag, dir)
	return dir
}

// makePathname creates a unique pathname for each test. It will panic if it
// cannot get the current working directory.
func makePathname() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	sep := string(filepath.Separator)
	return strings.Replace(p, sep, "_", -1)
}

// GetConfig returns a Tendermint config for the test cases.
func GetConfig() *tmcfg.Config {
	pathname := makePathname()
	config := tmcfg.ResetTestRoot(pathname)

	tmAddr, _, err := server.FreeTCPAddr()
	if err != nil {
		panic(err)
	}

	rcpAddr, _, err := server.FreeTCPAddr()
	if err != nil {
		panic(err)
	}

	grpcAddr, _, err := server.FreeTCPAddr()
	if err != nil {
		panic(err)
	}

	config.P2P.ListenAddress = tmAddr
	config.RPC.ListenAddress = rcpAddr
	config.RPC.GRPCListenAddress = grpcAddr

	return config
}
