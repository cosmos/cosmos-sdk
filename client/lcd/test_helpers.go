package lcd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	gapp "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	txbuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	tmrpc "github.com/tendermint/tendermint/rpc/lib/server"
	tmtypes "github.com/tendermint/tendermint/types"

	authRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	bankRest "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	govRest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	slashingRest "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	stakeRest "github.com/cosmos/cosmos-sdk/x/stake/client/rest"
)

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

// GetKeyBase returns the LCD test keybase. It also requires that a directory
// could be made and a keybase could be fetched.
//
// NOTE: memDB cannot be used because the request is expecting to interact with
// the default location.
func GetKeyBase(t *testing.T) crkeys.Keybase {
	dir, err := ioutil.TempDir("", "lcd_test")
	require.NoError(t, err)

	viper.Set(cli.HomeFlag, dir)

	keybase, err := keys.GetKeyBaseWithWritePerm()
	require.NoError(t, err)

	return keybase
}

// GetTestKeyBase fetches the current testing keybase
func GetTestKeyBase(t *testing.T) crkeys.Keybase {
	keybase, err := keys.GetKeyBaseWithWritePerm()
	require.NoError(t, err)
	return keybase
}

// CreateAddr adds an address to the key store and returns an address and seed.
// It also requires that the key could be created.
func CreateAddr(t *testing.T, name, password string, kb crkeys.Keybase) (sdk.AccAddress, string) {
	var (
		err  error
		info crkeys.Info
		seed string
	)

	info, seed, err = kb.CreateMnemonic(name, crkeys.English, password, crkeys.Secp256k1)
	require.NoError(t, err)

	return sdk.AccAddress(info.GetPubKey().Address()), seed
}

// Type that combines an Address with the pnemonic of the private key to that address
type AddrSeed struct {
	Address  sdk.AccAddress
	Seed     string
	Name     string
	Password string
}

// CreateAddr adds multiple address to the key store and returns the addresses and associated seeds in lexographical order by address.
// It also requires that the keys could be created.
func CreateAddrs(t *testing.T, kb crkeys.Keybase, numAddrs int) (addrs []sdk.AccAddress, seeds, names, passwords []string) {
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
		require.NoError(t, err)
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

// implement `Interface` in sort package.
type AddrSeedSlice []AddrSeed

func (b AddrSeedSlice) Len() int {
	return len(b)
}

// Sorts lexographically by Address
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

// InitializeTestLCD starts Tendermint and the LCD in process, listening on
// their respective sockets where nValidators is the total number of validators
// and initAddrs are the accounts to initialize with some steak tokens. It
// returns a cleanup function, a set of validator public keys, and a port.
func InitializeTestLCD(
	t *testing.T, nValidators int, initAddrs []sdk.AccAddress,
) (cleanup func(), valConsPubKeys []crypto.PubKey, valOperAddrs []sdk.ValAddress, port string) {

	if nValidators < 1 {
		panic("InitializeTestLCD must use at least one validator")
	}

	config := GetConfig()
	config.Consensus.TimeoutCommit = 100
	config.Consensus.SkipTimeoutCommit = false
	config.TxIndex.IndexAllTags = true

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger = log.NewFilter(logger, log.AllowError())

	privValidatorFile := config.PrivValidatorFile()
	privVal := pvm.LoadOrGenFilePV(privValidatorFile)
	privVal.Reset()

	db := dbm.NewMemDB()
	app := gapp.NewGaiaApp(logger, db, nil)
	cdc = gapp.MakeCodec()

	genesisFile := config.GenesisFile()
	genDoc, err := tmtypes.GenesisDocFromFile(genesisFile)
	require.Nil(t, err)
	genDoc.Validators = nil
	genDoc.SaveAs(genesisFile)
	genTxs := []json.RawMessage{}

	// append any additional (non-proposing) validators
	var accs []gapp.GenesisAccount
	for i := 0; i < nValidators; i++ {
		operPrivKey := secp256k1.GenPrivKey()
		operAddr := operPrivKey.PubKey().Address()
		pubKey := privVal.PubKey
		delegation := 100
		if i > 0 {
			pubKey = ed25519.GenPrivKey().PubKey()
			delegation = 1
		}
		msg := stake.NewMsgCreateValidator(
			sdk.ValAddress(operAddr),
			pubKey,
			sdk.NewCoin(stakeTypes.DefaultBondDenom, sdk.NewInt(int64(delegation))),
			stake.Description{Moniker: fmt.Sprintf("validator-%d", i+1)},
			stake.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		)
		stdSignMsg := txbuilder.StdSignMsg{
			ChainID: genDoc.ChainID,
			Msgs:    []sdk.Msg{msg},
		}
		sig, err := operPrivKey.Sign(stdSignMsg.Bytes())
		require.Nil(t, err)
		tx := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{{Signature: sig, PubKey: operPrivKey.PubKey()}}, "")
		txBytes, err := cdc.MarshalJSON(tx)
		require.Nil(t, err)
		genTxs = append(genTxs, txBytes)
		valConsPubKeys = append(valConsPubKeys, pubKey)
		valOperAddrs = append(valOperAddrs, sdk.ValAddress(operAddr))

		accAuth := auth.NewBaseAccountWithAddress(sdk.AccAddress(operAddr))
		accAuth.Coins = sdk.Coins{sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, 150)}
		accs = append(accs, gapp.NewGenesisAccount(&accAuth))
	}

	appGenState := gapp.NewDefaultGenesisState()
	appGenState.Accounts = accs
	genDoc.AppState, err = cdc.MarshalJSON(appGenState)
	require.NoError(t, err)
	genesisState, err := gapp.GaiaAppGenState(cdc, *genDoc, genTxs)
	require.NoError(t, err)

	// add some tokens to init accounts
	for _, addr := range initAddrs {
		accAuth := auth.NewBaseAccountWithAddress(addr)
		accAuth.Coins = sdk.Coins{sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, 100)}
		acc := gapp.NewGenesisAccount(&accAuth)
		genesisState.Accounts = append(genesisState.Accounts, acc)
		genesisState.StakeData.Pool.LooseTokens = genesisState.StakeData.Pool.LooseTokens.Add(sdk.NewDec(100))
	}

	appState, err := codec.MarshalJSONIndent(cdc, genesisState)
	require.NoError(t, err)
	genDoc.AppState = appState

	listenAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)

	// XXX: Need to set this so LCD knows the tendermint node address!
	viper.Set(client.FlagNode, config.RPC.ListenAddress)
	viper.Set(client.FlagChainID, genDoc.ChainID)
	// TODO Set to false once the upstream Tendermint proof verification issue is fixed.
	viper.Set(client.FlagTrustNode, true)
	dir, err := ioutil.TempDir("", "lcd_test")
	require.NoError(t, err)
	viper.Set(cli.HomeFlag, dir)

	node, err := startTM(config, logger, genDoc, privVal, app)
	require.NoError(t, err)

	tests.WaitForNextHeightTM(tests.ExtractPortFromAddress(config.RPC.ListenAddress))
	lcd, err := startLCD(logger, listenAddr, cdc, t)
	require.NoError(t, err)

	tests.WaitForLCDStart(port)
	tests.WaitForHeight(1, port)

	cleanup = func() {
		logger.Debug("cleaning up LCD initialization")
		node.Stop()
		node.Wait()
		lcd.Close()
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
func startLCD(logger log.Logger, listenAddr string, cdc *codec.Codec, t *testing.T) (net.Listener, error) {
	rs := NewRestServer(cdc)
	rs.setKeybase(GetTestKeyBase(t))
	registerRoutes(rs)
	listener, err := tmrpc.Listen(listenAddr, tmrpc.Config{})
	if err != nil {
		return nil, err
	}
	go tmrpc.StartHTTPServer(listener, rs.Mux, logger)
	return listener, nil
}

// NOTE: If making updates here also update cmd/gaia/cmd/gaiacli/main.go
func registerRoutes(rs *RestServer) {
	keys.RegisterRoutes(rs.Mux, rs.CliCtx.Indent)
	rpc.RegisterRoutes(rs.CliCtx, rs.Mux)
	tx.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
	authRest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, "acc")
	bankRest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	stakeRest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	slashingRest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	govRest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
}

// Request makes a test LCD test request. It returns a response object and a
// stringified response body.
func Request(t *testing.T, port, method, path string, payload []byte) (*http.Response, string) {
	var (
		err error
		res *http.Response
	)
	url := fmt.Sprintf("http://localhost:%v%v", port, path)
	fmt.Println("REQUEST " + method + " " + url)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	require.Nil(t, err)

	res, err = http.DefaultClient.Do(req)
	require.Nil(t, err)

	output, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.Nil(t, err)

	return res, string(output)
}
