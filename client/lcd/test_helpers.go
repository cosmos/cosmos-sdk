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

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	tmrpc "github.com/tendermint/tendermint/rpc/lib/server"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	keys "github.com/cosmos/cosmos-sdk/client/keys"
	gapp "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// f**ing long, but unique for each test
func makePathname() string {
	// get path
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	sep := string(filepath.Separator)
	return strings.Replace(p, sep, "_", -1)
}

// GetConfig returns a config for the test cases as a singleton
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

	config.P2P.ListenAddress = tmAddr
	config.RPC.ListenAddress = rcpAddr
	return config
}

// get the lcd test keybase
// note can't use a memdb because the request is expecting to interact with the default location
func GetKB(t *testing.T) crkeys.Keybase {
	dir, err := ioutil.TempDir("", "lcd_test")
	require.NoError(t, err)
	viper.Set(cli.HomeFlag, dir)
	keybase, err := keys.GetKeyBase() // dbm.NewMemDB()) // :(
	require.NoError(t, err)
	return keybase
}

// add an address to the store return name and password
func CreateAddr(t *testing.T, name, password string, kb crkeys.Keybase) (addr sdk.Address, seed string) {
	var info crkeys.Info
	var err error
	info, seed, err = kb.CreateMnemonic(name, crkeys.English, password, crkeys.Secp256k1)
	require.NoError(t, err)
	addr = info.GetPubKey().Address()
	return
}

// strt TM and the LCD in process, listening on their respective sockets
//   nValidators = number of validators
//   initAddrs = accounts to initialize with some steaks
func InitializeTestLCD(t *testing.T, nValidators int, initAddrs []sdk.Address) (cleanup func(), validatorsPKs []crypto.PubKey, port string) {

	config := GetConfig()
	config.Consensus.TimeoutCommit = 100
	config.Consensus.SkipTimeoutCommit = false
	config.TxIndex.IndexAllTags = true

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger = log.NewFilter(logger, log.AllowDebug())
	ctx := sdk.NewServerContext(config, logger)
	privValidatorFile := config.PrivValidatorFile()
	privVal := pvm.LoadOrGenFilePV(privValidatorFile)
	privVal.Reset()
	db := dbm.NewMemDB()
	app := gapp.NewGaiaApp(ctx, db)
	cdc = gapp.MakeCodec()

	genesisFile := config.GenesisFile()
	genDoc, err := tmtypes.GenesisDocFromFile(genesisFile)
	require.NoError(t, err)

	// add more validators
	if nValidators < 1 {
		panic("InitializeTestLCD must use at least one validator")
	}
	for i := 1; i < nValidators; i++ {
		genDoc.Validators = append(genDoc.Validators,
			tmtypes.GenesisValidator{
				PubKey: crypto.GenPrivKeyEd25519().PubKey(),
				Power:  1,
				Name:   "val",
			},
		)
	}

	// NOTE it's bad practice to reuse pk address for the owner address but doing in the
	// test for simplicity
	var appGenTxs []json.RawMessage
	for _, gdValidator := range genDoc.Validators {
		pk := gdValidator.PubKey
		validatorsPKs = append(validatorsPKs, pk) // append keys for output
		appGenTx, _, _, err := gapp.GaiaAppGenTxNF(cdc, pk, pk.Address(), "test_val1")
		require.NoError(t, err)
		appGenTxs = append(appGenTxs, appGenTx)
	}

	genesisState, err := gapp.GaiaAppGenState(cdc, appGenTxs[:])
	require.NoError(t, err)

	// add some tokens to init accounts
	for _, addr := range initAddrs {
		accAuth := auth.NewBaseAccountWithAddress(addr)
		accAuth.Coins = sdk.Coins{sdk.NewCoin("steak", 100)}
		acc := gapp.NewGenesisAccount(&accAuth)
		genesisState.Accounts = append(genesisState.Accounts, acc)
		genesisState.StakeData.Pool.LooseTokens += 100
	}

	appState, err := wire.MarshalJSONIndent(cdc, genesisState)
	require.NoError(t, err)
	genDoc.AppStateJSON = appState

	// LCD listen address
	var listenAddr string
	listenAddr, port, err = server.FreeTCPAddr()
	require.NoError(t, err)

	// XXX: need to set this so LCD knows the tendermint node address!
	viper.Set(client.FlagNode, config.RPC.ListenAddress)
	viper.Set(client.FlagChainID, genDoc.ChainID)

	node, err := startTM(config, logger, genDoc, privVal, app)
	require.NoError(t, err)
	lcd, err := startLCD(logger, listenAddr, cdc)
	require.NoError(t, err)

	//time.Sleep(time.Second)
	//tests.WaitForHeight(2, port)
	tests.WaitForLCDStart(port)
	tests.WaitForHeight(1, port)

	// for use in defer
	cleanup = func() {
		node.Stop()
		node.Wait()
		lcd.Close()
	}

	return
}

// Create & start in-process tendermint node with memdb
// and in-process abci application.
// TODO: need to clean up the WAL dir or enable it to be not persistent
func startTM(tmcfg *tmcfg.Config, logger log.Logger, genDoc *tmtypes.GenesisDoc, privVal tmtypes.PrivValidator, app abci.Application) (*nm.Node, error) {
	genDocProvider := func() (*tmtypes.GenesisDoc, error) { return genDoc, nil }
	dbProvider := func(*nm.DBContext) (dbm.DB, error) { return dbm.NewMemDB(), nil }
	n, err := nm.NewNode(tmcfg,
		privVal,
		proxy.NewLocalClientCreator(app),
		genDocProvider,
		dbProvider,
		nm.DefaultMetricsProvider,
		logger.With("module", "node"))
	if err != nil {
		return nil, err
	}

	err = n.Start()
	if err != nil {
		return nil, err
	}

	// wait for rpc
	tests.WaitForRPC(tmcfg.RPC.ListenAddress)

	logger.Info("Tendermint running!")
	return n, err
}

// start the LCD. note this blocks!
func startLCD(logger log.Logger, listenAddr string, cdc *wire.Codec) (net.Listener, error) {
	handler := createHandler(cdc)
	return tmrpc.StartHTTPServer(listenAddr, handler, logger, tmrpc.Config{})
}

// make a test lcd test request
func Request(t *testing.T, port, method, path string, payload []byte) (*http.Response, string) {
	var res *http.Response
	var err error
	url := fmt.Sprintf("http://localhost:%v%v", port, path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	require.Nil(t, err)
	res, err = http.DefaultClient.Do(req)
	//	res, err = http.Post(url, "application/json", bytes.NewBuffer(payload))
	require.Nil(t, err)

	output, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.Nil(t, err)

	return res, string(output)
}
