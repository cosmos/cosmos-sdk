package rest

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	gapp "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	amino "github.com/tendermint/go-amino"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	tmcfg "github.com/tendermint/tendermint/config"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	tmrpc "github.com/tendermint/tendermint/rpc/lib/server"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/cli"
)

func createConfig(t *testing.T) (config *tmcfg.Config, close func()) {
	dir, err := ioutil.TempDir("", "config")
	require.NoError(t, err)

	config = tmcfg.ResetTestRoot(dir)
	config.Consensus.SkipTimeoutCommit = true
	config.TxIndex.IndexAllTags = false

	p2pAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)
	config.P2P.ListenAddress = p2pAddr

	rpcAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)
	config.RPC.ListenAddress = rpcAddr

	close = func() {
		os.RemoveAll(dir)
	}

	return
}

func createGenesisDoc(t *testing.T, genesisFile string,
	codec *amino.Codec) *tmtypes.GenesisDoc {
	genesisDoc, err := tmtypes.GenesisDocFromFile(genesisFile)
	require.NoError(t, err)

	// NOTE it's bad practice to reuse pk address for the owner address but doing
	// in the test for simplicity
	var appGenTxs []json.RawMessage

	for _, validator := range genesisDoc.Validators {
		pubKey := validator.PubKey

		appGenTx, _, _, err := gapp.GaiaAppGenTxNF(codec, pubKey,
			sdk.AccAddress(pubKey.Address()), "moniker")

		require.NoError(t, err)
		appGenTxs = append(appGenTxs, appGenTx)
	}

	genesisState, err := gapp.GaiaAppGenState(codec, appGenTxs[:])
	require.NoError(t, err)

	appState, err := wire.MarshalJSONIndent(codec, genesisState)
	require.NoError(t, err)

	genesisDoc.AppState = appState
	viper.Set(client.FlagChainID, genesisDoc.ChainID)
	return genesisDoc
}

func createNode(t *testing.T, codec *amino.Codec, logger log.Logger,
) (reset func(), close func()) {
	config, closeConfig := createConfig(t)

	privValidator := pvm.LoadOrGenFilePV(config.PrivValidatorFile())
	privValidator.Reset()

	db := dbm.NewMemDB()
	app := gapp.NewGaiaApp(logger, db, nil)

	// XXX: need to set this so Gaia-Lite knows the tendermint node address!
	viper.Set(client.FlagNode, config.RPC.ListenAddress)

	dbProvider := func(*nm.DBContext) (dbm.DB, error) {
		return dbm.NewMemDB(), nil
	}

	genesisDocProvider := func() (*tmtypes.GenesisDoc, error) {
		return createGenesisDoc(t, config.GenesisFile(), codec), nil
	}

	node, err := nm.NewNode(
		config,
		privValidator,
		proxy.NewLocalClientCreator(app),
		genesisDocProvider,
		dbProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger.With("module", "node"))

	require.NoError(t, err)
	err = node.Start()
	require.NoError(t, err)
	tests.WaitForRPC(config.RPC.ListenAddress)

	reset = func() {
		tcmd.ResetAll(config.DBDir(), config.P2P.AddrBookFile(),
			config.PrivValidatorFile(), logger)
	}

	close = func() {
		node.Stop()
		node.Wait()
		closeConfig()
	}

	return
}

// a stripped-down version of the Gaia-Lite handler
func createHandler(t *testing.T, codec *wire.Codec) http.Handler {
	cliCtx := context.NewCLIContext().WithCodec(codec).WithLogger(os.Stdout)
	router := mux.NewRouter()
	rpc.RegisterRoutes(cliCtx, router)

	dir, err := ioutil.TempDir("", "lcd_test")
	require.NoError(t, err)

	viper.Set(cli.HomeFlag, dir)

	keyBase, err := keys.GetKeyBase()
	require.NoError(t, err)

	RegisterRoutes(cliCtx, router, codec, keyBase)
	return router
}

func createGaiaLite(t *testing.T, codec *amino.Codec, logger log.Logger) (gaiaLite net.Listener,
	port string) {
	listenAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)

	handler := createHandler(t, codec)

	gaiaLite, err = tmrpc.StartHTTPServer(listenAddr, handler, logger,
		tmrpc.Config{})

	require.NoError(t, err)
	tests.WaitForLCDStart(port)
	tests.WaitForHeight(1, port)
	return
}

func createTestNetwork(t *testing.T) (url string, reset func(), close func()) {
	codec := gapp.MakeCodec()

	unfiltered := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger := log.NewFilter(unfiltered, log.AllowDebug())

	reset, nodeClose := createNode(t, codec, logger)
	gaiaLite, port := createGaiaLite(t, codec, logger)
	url = "http://localhost:" + port

	close = func() {
		nodeClose()
		gaiaLite.Close()
	}

	return
}

// Create a separate HTTP server to handle Pact state setup requests.
func createSetupServer(t *testing.T, reset func()) string {
	_, port, err := server.FreeTCPAddr()
	require.NoError(t, err)

	handler := http.HandlerFunc(func(writer http.ResponseWriter,
		request *http.Request) {
		// Reset the state of the Gaiad node to genesis so that each test is
		// indepedent.
		reset()

		var state *types.ProviderState
		decoder := json.NewDecoder(request.Body)
		decoder.Decode(&state)

		if state.State == "delegated" {
			// We don't do anything yet.
		}

		writer.Header().Set("Content-Type", "application/json")
	})

	go http.ListenAndServe(":"+port, handler)
	return "http://localhost:" + port
}
func TestProvider(t *testing.T) {
	// Create Pact connecting to local Daemon
	pact := &dsl.Pact{
		Consumer: "Voyager",
		Provider: "LCD",
	}

	// Start provider API in the background
	gaiaLiteURL, reset, close := createTestNetwork(t)
	defer close()

	// Verify the Provider with local Pact Files
	pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:        gaiaLiteURL,
		PactURLs:               []string{filepath.ToSlash("voyager-cosmos-lite.json")},
		ProviderStatesSetupURL: createSetupServer(t, reset),
	})
}
