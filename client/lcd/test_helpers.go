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
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	gapp "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	tmrpc "github.com/tendermint/tendermint/rpc/lib/server"
	tmtypes "github.com/tendermint/tendermint/types"
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

	keybase, err := keys.GetKeyBase()
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

// InitializeTestLCD starts Tendermint and the LCD in process, listening on
// their respective sockets where nValidators is the total number of validators
// and initAddrs are the accounts to initialize with some steak tokens. It
// returns a cleanup function, a set of validator public keys, and a port.
func InitializeTestLCD(t *testing.T, nValidators int, initAddrs []sdk.AccAddress) (func(), []crypto.PubKey, string) {
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
	require.NoError(t, err)

	if nValidators < 1 {
		panic("InitializeTestLCD must use at least one validator")
	}

	for i := 1; i < nValidators; i++ {
		genDoc.Validators = append(genDoc.Validators,
			tmtypes.GenesisValidator{
				PubKey: ed25519.GenPrivKey().PubKey(),
				Power:  1,
				Name:   "val",
			},
		)
	}

	var validatorsPKs []crypto.PubKey

	// NOTE: It's bad practice to reuse public key address for the owner
	// address but doing in the test for simplicity.
	var appGenTxs []json.RawMessage
	for _, gdValidator := range genDoc.Validators {
		pk := gdValidator.PubKey
		validatorsPKs = append(validatorsPKs, pk)

		appGenTx, _, _, err := gapp.GaiaAppGenTxNF(cdc, pk, sdk.AccAddress(pk.Address()), "test_val1")
		require.NoError(t, err)

		appGenTxs = append(appGenTxs, appGenTx)
	}

	genesisState, err := gapp.GaiaAppGenState(cdc, appGenTxs[:])
	require.NoError(t, err)

	// add some tokens to init accounts
	for _, addr := range initAddrs {
		accAuth := auth.NewBaseAccountWithAddress(addr)
		accAuth.Coins = sdk.Coins{sdk.NewInt64Coin("steak", 100)}
		acc := gapp.NewGenesisAccount(&accAuth)
		genesisState.Accounts = append(genesisState.Accounts, acc)
		genesisState.StakeData.Pool.LooseTokens = genesisState.StakeData.Pool.LooseTokens.Add(sdk.NewRat(100))
	}

	appState, err := wire.MarshalJSONIndent(cdc, genesisState)
	require.NoError(t, err)
	genDoc.AppState = appState

	listenAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)

	// XXX: Need to set this so LCD knows the tendermint node address!
	viper.Set(client.FlagNode, config.RPC.ListenAddress)
	viper.Set(client.FlagChainID, genDoc.ChainID)

	node, err := startTM(config, logger, genDoc, privVal, app)
	require.NoError(t, err)

	lcd, err := startLCD(logger, listenAddr, cdc)
	require.NoError(t, err)

	tests.WaitForLCDStart(port)
	tests.WaitForHeight(1, port)

	cleanup := func() {
		logger.Debug("cleaning up LCD initialization")
		node.Stop()
		node.Wait()
		lcd.Close()
	}

	return cleanup, validatorsPKs, port
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
	node, err := nm.NewNode(
		tmcfg,
		privVal,
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
//
// NOTE: This causes the thread to block.
func startLCD(logger log.Logger, listenAddr string, cdc *wire.Codec) (net.Listener, error) {
	return tmrpc.StartHTTPServer(listenAddr, createHandler(cdc), logger, tmrpc.Config{})
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
