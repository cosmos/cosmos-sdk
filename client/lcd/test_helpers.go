package lcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

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
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	txbuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	bankrest "github.com/cosmos/cosmos-sdk/x/bank/client/rest"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrrest "github.com/cosmos/cosmos-sdk/x/distribution/client/rest"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govrest "github.com/cosmos/cosmos-sdk/x/gov/client/rest"
	gcutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingrest "github.com/cosmos/cosmos-sdk/x/slashing/client/rest"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingrest "github.com/cosmos/cosmos-sdk/x/staking/client/rest"

	abci "github.com/tendermint/tendermint/abci/types"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
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
func InitClientHome(t *testing.T, dir string) string {
	var err error
	if dir == "" {
		dir, err = ioutil.TempDir("", "lcd_test")
		require.NoError(t, err)
	}
	// TODO: this should be set in NewRestServer
	// and pass down the CLIContext to achieve
	// parallelism.
	viper.Set(cli.HomeFlag, dir)
	return dir
}

// TODO: Make InitializeTestLCD safe to call in multiple tests at the same time
// InitializeTestLCD starts Tendermint and the LCD in process, listening on
// their respective sockets where nValidators is the total number of validators
// and initAddrs are the accounts to initialize with some steak tokens. It
// returns a cleanup function, a set of validator public keys, and a port.
func InitializeTestLCD(t *testing.T, nValidators int, initAddrs []sdk.AccAddress, minting bool) (
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
	app := gapp.NewGaiaApp(logger, db, nil, true)
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
		require.Nil(t, err)
		tx := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{{Signature: sig, PubKey: operPrivKey.PubKey()}}, "")
		txBytes, err := cdc.MarshalJSON(tx)
		require.Nil(t, err)

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
	require.NoError(t, err)
	genesisState, err := gapp.GaiaAppGenState(cdc, *genDoc, genTxs)
	require.NoError(t, err)

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

	// double check inflation is set according to the minting boolean flag
	if minting {
		require.Equal(t, sdk.MustNewDecFromStr("15000.0"),
			genesisState.MintData.Params.InflationMax)
		require.Equal(t, sdk.MustNewDecFromStr("10000.0"), genesisState.MintData.Minter.Inflation)
		require.Equal(t, sdk.MustNewDecFromStr("10000.0"),
			genesisState.MintData.Params.InflationMin)
	} else {
		require.Equal(t, sdk.ZeroDec(), genesisState.MintData.Params.InflationMax)
		require.Equal(t, sdk.ZeroDec(), genesisState.MintData.Minter.Inflation)
		require.Equal(t, sdk.ZeroDec(), genesisState.MintData.Params.InflationMin)
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
	authrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, auth.StoreKey)
	bankrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	distrrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, distr.StoreKey)
	stakingrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	slashingrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc, rs.KeyBase)
	govrest.RegisterRoutes(rs.CliCtx, rs.Mux, rs.Cdc)
}

// Request makes a test LCD test request. It returns a response object and a
// stringified response body.
func Request(t *testing.T, port, method, path string, payload []byte) (*http.Response, string) {
	var (
		err error
		res *http.Response
	)
	url := fmt.Sprintf("http://localhost:%v%v", port, path)
	fmt.Printf("REQUEST %s %s\n", method, url)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	require.Nil(t, err)

	res, err = http.DefaultClient.Do(req)
	require.Nil(t, err)

	output, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.Nil(t, err)

	return res, string(output)
}

// ----------------------------------------------------------------------
// ICS 0 - Tendermint
// ----------------------------------------------------------------------
// GET /node_info The properties of the connected node
func getNodeInfo(t *testing.T, port string) p2p.DefaultNodeInfo {
	res, body := Request(t, port, "GET", "/node_info", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var nodeInfo p2p.DefaultNodeInfo
	err := cdc.UnmarshalJSON([]byte(body), &nodeInfo)
	require.Nil(t, err, "Couldn't parse node info")

	require.NotEqual(t, p2p.DefaultNodeInfo{}, nodeInfo, "res: %v", res)
	return nodeInfo
}

// GET /syncing Syncing state of node
func getSyncStatus(t *testing.T, port string, syncing bool) {
	res, body := Request(t, port, "GET", "/syncing", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	if syncing {
		require.Equal(t, "true", body)
		return
	}
	require.Equal(t, "false", body)
}

// GET /blocks/latest Get the latest block
// GET /blocks/{height} Get a block at a certain height
func getBlock(t *testing.T, port string, height int, expectFail bool) ctypes.ResultBlock {
	var url string
	if height > 0 {
		url = fmt.Sprintf("/blocks/%d", height)
	} else {
		url = "/blocks/latest"
	}
	var resultBlock ctypes.ResultBlock

	res, body := Request(t, port, "GET", url, nil)
	if expectFail {
		require.Equal(t, http.StatusNotFound, res.StatusCode, body)
		return resultBlock
	}
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultBlock)
	require.Nil(t, err, "Couldn't parse block")

	require.NotEqual(t, ctypes.ResultBlock{}, resultBlock)
	return resultBlock
}

// GET /validatorsets/{height} Get a validator set a certain height
// GET /validatorsets/latest Get the latest validator set
func getValidatorSets(t *testing.T, port string, height int, expectFail bool) rpc.ResultValidatorsOutput {
	var url string
	if height > 0 {
		url = fmt.Sprintf("/validatorsets/%d", height)
	} else {
		url = "/validatorsets/latest"
	}
	var resultVals rpc.ResultValidatorsOutput

	res, body := Request(t, port, "GET", url, nil)

	if expectFail {
		require.Equal(t, http.StatusNotFound, res.StatusCode, body)
		return resultVals
	}

	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultVals)
	require.Nil(t, err, "Couldn't parse validatorset")

	require.NotEqual(t, rpc.ResultValidatorsOutput{}, resultVals)
	return resultVals
}

// GET /txs/{hash} get tx by hash
func getTransaction(t *testing.T, port string, hash string) sdk.TxResponse {
	var tx sdk.TxResponse
	res, body := getTransactionRequest(t, port, hash)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &tx)
	require.NoError(t, err)
	return tx
}

func getTransactionRequest(t *testing.T, port, hash string) (*http.Response, string) {
	return Request(t, port, "GET", fmt.Sprintf("/txs/%s", hash), nil)
}

// POST /txs broadcast txs

// GET /txs search transactions
func getTransactions(t *testing.T, port string, tags ...string) []sdk.TxResponse {
	var txs []sdk.TxResponse
	if len(tags) == 0 {
		return txs
	}
	queryStr := strings.Join(tags, "&")
	res, body := Request(t, port, "GET", fmt.Sprintf("/txs?%s", queryStr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &txs)
	require.NoError(t, err)
	return txs
}

// ----------------------------------------------------------------------
// ICS 1 - Keys
// ----------------------------------------------------------------------
// GET /keys List of accounts stored locally
func getKeys(t *testing.T, port string) []keys.KeyOutput {
	res, body := Request(t, port, "GET", "/keys", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var m []keys.KeyOutput
	err := cdc.UnmarshalJSON([]byte(body), &m)
	require.Nil(t, err)
	return m
}

// POST /keys Create a new account locally
func doKeysPost(t *testing.T, port, name, password, mnemonic string, account int, index int) keys.KeyOutput {
	pk := keys.AddNewKey{name, password, mnemonic, account, index}
	req, err := cdc.MarshalJSON(pk)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", "/keys", req)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err, body)
	return resp
}

// GET /keys/seed Create a new seed to create a new account defaultValidFor
func getKeysSeed(t *testing.T, port string) string {
	res, body := Request(t, port, "GET", "/keys/seed", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	reg, err := regexp.Compile(`([a-z]+ ){12}`)
	require.Nil(t, err)
	match := reg.MatchString(body)
	require.True(t, match, "Returned seed has wrong format", body)
	return body
}

// POST /keys/{name}/recove Recover a account from a seed
func doRecoverKey(t *testing.T, port, recoverName, recoverPassword, mnemonic string, account uint32, index uint32) {
	pk := keys.RecoverKey{recoverPassword, mnemonic, int(account), int(index)}
	req, err := cdc.MarshalJSON(pk)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/keys/%s/recover", recoverName), req)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err = codec.Cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err, body)

	addr1Bech32 := resp.Address
	_, err = sdk.AccAddressFromBech32(addr1Bech32)
	require.NoError(t, err, "Failed to return a correct bech32 address")
}

// GET /keys/{name} Get a certain locally stored account
func getKey(t *testing.T, port, name string) keys.KeyOutput {
	res, body := Request(t, port, "GET", fmt.Sprintf("/keys/%s", name), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err := cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err)
	return resp
}

// PUT /keys/{name} Update the password for this account in the KMS
func updateKey(t *testing.T, port, name, oldPassword, newPassword string, fail bool) {
	kr := keys.UpdateKeyReq{oldPassword, newPassword}
	req, err := cdc.MarshalJSON(kr)
	require.NoError(t, err)
	keyEndpoint := fmt.Sprintf("/keys/%s", name)
	res, body := Request(t, port, "PUT", keyEndpoint, req)
	if fail {
		require.Equal(t, http.StatusUnauthorized, res.StatusCode, body)
		return
	}
	require.Equal(t, http.StatusOK, res.StatusCode, body)
}

// DELETE /keys/{name} Remove an account
func deleteKey(t *testing.T, port, name, password string) {
	dk := keys.DeleteKeyReq{password}
	req, err := cdc.MarshalJSON(dk)
	require.NoError(t, err)
	keyEndpoint := fmt.Sprintf("/keys/%s", name)
	res, body := Request(t, port, "DELETE", keyEndpoint, req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
}

// GET /auth/accounts/{address} Get the account information on blockchain
func getAccount(t *testing.T, port string, addr sdk.AccAddress) auth.Account {
	res, body := Request(t, port, "GET", fmt.Sprintf("/auth/accounts/%s", addr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var acc auth.Account
	err := cdc.UnmarshalJSON([]byte(body), &acc)
	require.Nil(t, err)
	return acc
}

// ----------------------------------------------------------------------
// ICS 20 - Tokens
// ----------------------------------------------------------------------

// POST /tx/sign Sign a Tx
func doSign(t *testing.T, port, name, password, chainID string, accnum, sequence uint64, msg auth.StdTx) auth.StdTx {
	var signedMsg auth.StdTx
	payload := authrest.SignBody{
		Tx: msg,
		BaseReq: rest.NewBaseReq(
			name, password, "", chainID, "", "", accnum, sequence, nil, nil, false, false,
		),
	}
	json, err := cdc.MarshalJSON(payload)
	require.Nil(t, err)
	res, body := Request(t, port, "POST", "/tx/sign", json)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &signedMsg))
	return signedMsg
}

// POST /tx/broadcast Send a signed Tx
func doBroadcast(t *testing.T, port string, msg auth.StdTx) sdk.TxResponse {
	tx := authrest.BroadcastReq{Tx: msg, Return: "block"}
	req, err := cdc.MarshalJSON(tx)
	require.Nil(t, err)
	res, body := Request(t, port, "POST", "/tx/broadcast", req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resultTx sdk.TxResponse
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &resultTx))
	return resultTx
}

// GET /bank/balances/{address} Get the account balances

// POST /bank/accounts/{address}/transfers Send coins (build -> sign -> send)
func doTransfer(t *testing.T, port, seed, name, memo, password string, addr sdk.AccAddress, fees sdk.Coins) (receiveAddr sdk.AccAddress, resultTx sdk.TxResponse) {
	res, body, receiveAddr := doTransferWithGas(t, port, seed, name, memo, password, addr, "", 1.0, false, false, fees)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return receiveAddr, resultTx
}

func doTransferWithGas(
	t *testing.T, port, seed, from, memo, password string, addr sdk.AccAddress,
	gas string, gasAdjustment float64, simulate, generateOnly bool, fees sdk.Coins,
) (res *http.Response, body string, receiveAddr sdk.AccAddress) {

	// create receive address
	kb := crkeys.NewInMemory()

	receiveInfo, _, err := kb.CreateMnemonic(
		"receive_address", crkeys.English, gapp.DefaultKeyPass, crkeys.SigningAlgo("secp256k1"),
	)
	require.Nil(t, err)

	receiveAddr = sdk.AccAddress(receiveInfo.GetPubKey().Address())
	acc := getAccount(t, port, addr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)

	if generateOnly {
		// generate only txs do not use a Keybase so the address must be used
		from = addr.String()
	}

	baseReq := rest.NewBaseReq(
		from, password, memo, chainID, gas,
		fmt.Sprintf("%f", gasAdjustment), accnum, sequence, fees, nil,
		generateOnly, simulate,
	)

	sr := bankrest.SendReq{
		Amount:  sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)},
		BaseReq: baseReq,
	}

	req, err := cdc.MarshalJSON(sr)
	require.NoError(t, err)

	res, body = Request(t, port, "POST", fmt.Sprintf("/bank/accounts/%s/transfers", receiveAddr), req)
	return
}

func doTransferWithGasAccAuto(
	t *testing.T, port, seed, from, memo, password string, gas string,
	gasAdjustment float64, simulate, generateOnly bool, fees sdk.Coins,
) (res *http.Response, body string, receiveAddr sdk.AccAddress) {

	// create receive address
	kb := crkeys.NewInMemory()

	receiveInfo, _, err := kb.CreateMnemonic(
		"receive_address", crkeys.English, gapp.DefaultKeyPass, crkeys.SigningAlgo("secp256k1"),
	)
	require.Nil(t, err)

	receiveAddr = sdk.AccAddress(receiveInfo.GetPubKey().Address())
	chainID := viper.GetString(client.FlagChainID)

	baseReq := rest.NewBaseReq(
		from, password, memo, chainID, gas,
		fmt.Sprintf("%f", gasAdjustment), 0, 0, fees, nil, generateOnly, simulate,
	)

	sr := bankrest.SendReq{
		Amount:  sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)},
		BaseReq: baseReq,
	}

	req, err := cdc.MarshalJSON(sr)
	require.NoError(t, err)

	res, body = Request(t, port, "POST", fmt.Sprintf("/bank/accounts/%s/transfers", receiveAddr), req)
	return
}

// ----------------------------------------------------------------------
// ICS 21 - Stake
// ----------------------------------------------------------------------

// POST /staking/delegators/{delegatorAddr}/delegations Submit delegation
func doDelegate(t *testing.T, port, name, password string,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Int, fees sdk.Coins) (resultTx sdk.TxResponse) {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	baseReq := rest.NewBaseReq(name, password, "", chainID, "", "", accnum, sequence, fees, nil, false, false)
	msg := msgDelegationsInput{
		BaseReq:       baseReq,
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		Delegation:    sdk.NewCoin(sdk.DefaultBondDenom, amount),
	}
	req, err := cdc.MarshalJSON(msg)
	require.NoError(t, err)
	res, body := Request(t, port, "POST", fmt.Sprintf("/staking/delegators/%s/delegations", delAddr.String()), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var result sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &result)
	require.Nil(t, err)

	return result
}

type msgDelegationsInput struct {
	BaseReq       rest.BaseReq   `json:"base_req"`
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
	ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
	Delegation    sdk.Coin       `json:"delegation"`
}

// POST /staking/delegators/{delegatorAddr}/delegations Submit delegation
func doUndelegate(t *testing.T, port, name, password string,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Int, fees sdk.Coins) (resultTx sdk.TxResponse) {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	baseReq := rest.NewBaseReq(name, password, "", chainID, "", "", accnum, sequence, fees, nil, false, false)
	msg := msgUndelegateInput{
		BaseReq:       baseReq,
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		SharesAmount:  sdk.NewDecFromInt(amount),
	}
	req, err := cdc.MarshalJSON(msg)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/staking/delegators/%s/unbonding_delegations", delAddr), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var result sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &result)
	require.Nil(t, err)

	return result
}

type msgUndelegateInput struct {
	BaseReq       rest.BaseReq   `json:"base_req"`
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
	ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
	SharesAmount  sdk.Dec        `json:"shares"`
}

// POST /staking/delegators/{delegatorAddr}/delegations Submit delegation
func doBeginRedelegation(t *testing.T, port, name, password string,
	delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress, amount sdk.Int,
	fees sdk.Coins) (resultTx sdk.TxResponse) {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)
	baseReq := rest.NewBaseReq(name, password, "", chainID, "", "", accnum, sequence, fees, nil, false, false)

	msg := stakingrest.MsgBeginRedelegateInput{
		BaseReq:          baseReq,
		DelegatorAddr:    delAddr,
		ValidatorSrcAddr: valSrcAddr,
		ValidatorDstAddr: valDstAddr,
		SharesAmount:     sdk.NewDecFromInt(amount),
	}
	req, err := cdc.MarshalJSON(msg)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/staking/delegators/%s/redelegations", delAddr), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var result sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &result)
	require.Nil(t, err)

	return result
}

type msgBeginRedelegateInput struct {
	BaseReq          rest.BaseReq   `json:"base_req"`
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`     // in bech32
	ValidatorSrcAddr sdk.ValAddress `json:"validator_src_addr"` // in bech32
	ValidatorDstAddr sdk.ValAddress `json:"validator_dst_addr"` // in bech32
	SharesAmount     sdk.Dec        `json:"shares"`
}

// GET /staking/delegators/{delegatorAddr}/delegations Get all delegations from a delegator
func getDelegatorDelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) []staking.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/delegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var dels []staking.Delegation

	err := cdc.UnmarshalJSON([]byte(body), &dels)
	require.Nil(t, err)

	return dels
}

// GET /staking/delegators/{delegatorAddr}/unbonding_delegations Get all unbonding delegations from a delegator
func getDelegatorUnbondingDelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) []staking.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/unbonding_delegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var ubds []staking.UnbondingDelegation

	err := cdc.UnmarshalJSON([]byte(body), &ubds)
	require.Nil(t, err)

	return ubds
}

// GET /staking/redelegations?delegator=0xdeadbeef&validator_from=0xdeadbeef&validator_to=0xdeadbeef& Get redelegations filters by params passed in
func getRedelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress, srcValidatorAddr sdk.ValAddress, dstValidatorAddr sdk.ValAddress) []staking.Redelegation {
	var res *http.Response
	var body string
	endpoint := "/staking/redelegations?"
	if !delegatorAddr.Empty() {
		endpoint += fmt.Sprintf("delegator=%s&", delegatorAddr)
	}
	if !srcValidatorAddr.Empty() {
		endpoint += fmt.Sprintf("validator_from=%s&", srcValidatorAddr)
	}
	if !dstValidatorAddr.Empty() {
		endpoint += fmt.Sprintf("validator_to=%s&", dstValidatorAddr)
	}
	res, body = Request(t, port, "GET", endpoint, nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var redels []staking.Redelegation
	err := cdc.UnmarshalJSON([]byte(body), &redels)
	require.Nil(t, err)
	return redels
}

// GET /staking/delegators/{delegatorAddr}/validators Query all validators that a delegator is bonded to
func getDelegatorValidators(t *testing.T, port string, delegatorAddr sdk.AccAddress) []staking.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/validators", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bondedValidators []staking.Validator

	err := cdc.UnmarshalJSON([]byte(body), &bondedValidators)
	require.Nil(t, err)

	return bondedValidators
}

// GET /staking/delegators/{delegatorAddr}/validators/{validatorAddr} Query a validator that a delegator is bonded to
func getDelegatorValidator(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) staking.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/validators/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bondedValidator staking.Validator
	err := cdc.UnmarshalJSON([]byte(body), &bondedValidator)
	require.Nil(t, err)

	return bondedValidator
}

// GET /staking/delegators/{delegatorAddr}/txs Get all staking txs (i.e msgs) from a delegator
func getBondingTxs(t *testing.T, port string, delegatorAddr sdk.AccAddress, query string) []sdk.TxResponse {
	var res *http.Response
	var body string

	if len(query) > 0 {
		res, body = Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/txs?type=%s", delegatorAddr, query), nil)
	} else {
		res, body = Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/txs", delegatorAddr), nil)
	}
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var txs []sdk.TxResponse

	err := cdc.UnmarshalJSON([]byte(body), &txs)
	require.Nil(t, err)

	return txs
}

// GET /staking/delegators/{delegatorAddr}/delegations/{validatorAddr} Query the current delegation between a delegator and a validator
func getDelegation(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) staking.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/delegators/%s/delegations/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bond staking.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &bond)
	require.Nil(t, err)

	return bond
}

// GET /staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr} Query all unbonding delegations between a delegator and a validator
func getUnbondingDelegation(t *testing.T, port string, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress) staking.UnbondingDelegation {

	res, body := Request(t, port, "GET",
		fmt.Sprintf("/staking/delegators/%s/unbonding_delegations/%s",
			delegatorAddr, validatorAddr), nil)

	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var unbond staking.UnbondingDelegation
	err := cdc.UnmarshalJSON([]byte(body), &unbond)
	require.Nil(t, err)

	return unbond
}

// GET /staking/validators Get all validator candidates
func getValidators(t *testing.T, port string) []staking.Validator {
	res, body := Request(t, port, "GET", "/staking/validators", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var validators []staking.Validator
	err := cdc.UnmarshalJSON([]byte(body), &validators)
	require.Nil(t, err)

	return validators
}

// GET /staking/validators/{validatorAddr} Query the information from a single validator
func getValidator(t *testing.T, port string, validatorAddr sdk.ValAddress) staking.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/validators/%s", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var validator staking.Validator
	err := cdc.UnmarshalJSON([]byte(body), &validator)
	require.Nil(t, err)

	return validator
}

// GET /staking/validators/{validatorAddr}/delegations Get all delegations from a validator
func getValidatorDelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []staking.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/validators/%s/delegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var delegations []staking.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &delegations)
	require.Nil(t, err)

	return delegations
}

// GET /staking/validators/{validatorAddr}/unbonding_delegations Get all unbonding delegations from a validator
func getValidatorUnbondingDelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []staking.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/staking/validators/%s/unbonding_delegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var ubds []staking.UnbondingDelegation
	err := cdc.UnmarshalJSON([]byte(body), &ubds)
	require.Nil(t, err)

	return ubds
}

// GET /staking/pool Get the current state of the staking pool
func getStakingPool(t *testing.T, port string) staking.Pool {
	res, body := Request(t, port, "GET", "/staking/pool", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NotNil(t, body)
	var pool staking.Pool
	err := cdc.UnmarshalJSON([]byte(body), &pool)
	require.Nil(t, err)
	return pool
}

// GET /staking/parameters Get the current staking parameter values
func getStakingParams(t *testing.T, port string) staking.Params {
	res, body := Request(t, port, "GET", "/staking/parameters", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var params staking.Params
	err := cdc.UnmarshalJSON([]byte(body), &params)
	require.Nil(t, err)
	return params
}

// ----------------------------------------------------------------------
// ICS 22 - Gov
// ----------------------------------------------------------------------
// POST /gov/proposals Submit a proposal
func doSubmitProposal(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress,
	amount sdk.Int, fees sdk.Coins) (resultTx sdk.TxResponse) {

	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	baseReq := rest.NewBaseReq(name, password, "", chainID, "", "", accnum, sequence, fees, nil, false, false)

	pr := govrest.PostProposalReq{
		Title:          "Test",
		Description:    "test",
		ProposalType:   "Text",
		Proposer:       proposerAddr,
		InitialDeposit: sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, amount)},
		BaseReq:        baseReq,
	}

	req, err := cdc.MarshalJSON(pr)
	require.NoError(t, err)

	// submitproposal
	res, body := Request(t, port, "POST", "/gov/proposals", req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

// GET /gov/proposals Query proposals
func getProposalsAll(t *testing.T, port string) []gov.Proposal {
	res, body := Request(t, port, "GET", "/gov/proposals", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

// GET /gov/proposals?depositor=%s Query proposals
func getProposalsFilterDepositor(t *testing.T, port string, depositorAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositor=%s", depositorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

// GET /gov/proposals?voter=%s Query proposals
func getProposalsFilterVoter(t *testing.T, port string, voterAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?voter=%s", voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

// GET /gov/proposals?depositor=%s&voter=%s Query proposals
func getProposalsFilterVoterDepositor(t *testing.T, port string, voterAddr, depositorAddr sdk.AccAddress) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?depositor=%s&voter=%s", depositorAddr, voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

// GET /gov/proposals?status=%s Query proposals
func getProposalsFilterStatus(t *testing.T, port string, status gov.ProposalStatus) []gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals?status=%s", status), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposals []gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposals)
	require.Nil(t, err)
	return proposals
}

// POST /gov/proposals/{proposalId}/deposits Deposit tokens to a proposal
func doDeposit(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, proposalID uint64,
	amount sdk.Int, fees sdk.Coins) (resultTx sdk.TxResponse) {

	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	baseReq := rest.NewBaseReq(name, password, "", chainID, "", "", accnum, sequence, fees, nil, false, false)

	dr := govrest.DepositReq{
		Depositor: proposerAddr,
		Amount:    sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, amount)},
		BaseReq:   baseReq,
	}

	req, err := cdc.MarshalJSON(dr)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/deposits", proposalID), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

// GET /gov/proposals/{proposalId}/deposits Query deposits
func getDeposits(t *testing.T, port string, proposalID uint64) []gov.Deposit {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/deposits", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var deposits []gov.Deposit
	err := cdc.UnmarshalJSON([]byte(body), &deposits)
	require.Nil(t, err)
	return deposits
}

// GET /gov/proposals/{proposalId}/tally Get a proposal's tally result at the current time
func getTally(t *testing.T, port string, proposalID uint64) gov.TallyResult {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/tally", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var tally gov.TallyResult
	err := cdc.UnmarshalJSON([]byte(body), &tally)
	require.Nil(t, err)
	return tally
}

// POST /gov/proposals/{proposalId}/votes Vote a proposal
func doVote(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, proposalID uint64, option string, fees sdk.Coins) (resultTx sdk.TxResponse) {
	// get the account to get the sequence
	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	baseReq := rest.NewBaseReq(name, password, "", chainID, "", "", accnum, sequence, fees, nil, false, false)

	vr := govrest.VoteReq{
		Voter:   proposerAddr,
		Option:  option,
		BaseReq: baseReq,
	}

	req, err := cdc.MarshalJSON(vr)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

// GET /gov/proposals/{proposalId}/votes Query voters
func getVotes(t *testing.T, port string, proposalID uint64) []gov.Vote {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var votes []gov.Vote
	err := cdc.UnmarshalJSON([]byte(body), &votes)
	require.Nil(t, err)
	return votes
}

// GET /gov/proposals/{proposalId} Query a proposal
func getProposal(t *testing.T, port string, proposalID uint64) gov.Proposal {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var proposal gov.Proposal
	err := cdc.UnmarshalJSON([]byte(body), &proposal)
	require.Nil(t, err)
	return proposal
}

// GET /gov/proposals/{proposalId}/deposits/{depositor} Query deposit
func getDeposit(t *testing.T, port string, proposalID uint64, depositorAddr sdk.AccAddress) gov.Deposit {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/deposits/%s", proposalID, depositorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var deposit gov.Deposit
	err := cdc.UnmarshalJSON([]byte(body), &deposit)
	require.Nil(t, err)
	return deposit
}

// GET /gov/proposals/{proposalId}/votes/{voter} Query vote
func getVote(t *testing.T, port string, proposalID uint64, voterAddr sdk.AccAddress) gov.Vote {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/votes/%s", proposalID, voterAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var vote gov.Vote
	err := cdc.UnmarshalJSON([]byte(body), &vote)
	require.Nil(t, err)
	return vote
}

// GET /gov/proposals/{proposalId}/proposer
func getProposer(t *testing.T, port string, proposalID uint64) gcutils.Proposer {
	res, body := Request(t, port, "GET", fmt.Sprintf("/gov/proposals/%d/proposer", proposalID), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var proposer gcutils.Proposer
	err := cdc.UnmarshalJSON([]byte(body), &proposer)

	require.Nil(t, err)
	return proposer
}

// GET /gov/parameters/deposit Query governance deposit parameters
func getDepositParam(t *testing.T, port string) gov.DepositParams {
	res, body := Request(t, port, "GET", "/gov/parameters/deposit", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var depositParams gov.DepositParams
	err := cdc.UnmarshalJSON([]byte(body), &depositParams)
	require.Nil(t, err)
	return depositParams
}

// GET /gov/parameters/tallying Query governance tally parameters
func getTallyingParam(t *testing.T, port string) gov.TallyParams {
	res, body := Request(t, port, "GET", "/gov/parameters/tallying", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var tallyParams gov.TallyParams
	err := cdc.UnmarshalJSON([]byte(body), &tallyParams)
	require.Nil(t, err)
	return tallyParams
}

// GET /gov/parameters/voting Query governance voting parameters
func getVotingParam(t *testing.T, port string) gov.VotingParams {
	res, body := Request(t, port, "GET", "/gov/parameters/voting", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var votingParams gov.VotingParams
	err := cdc.UnmarshalJSON([]byte(body), &votingParams)
	require.Nil(t, err)
	return votingParams
}

// ----------------------------------------------------------------------
// ICS 23 - Slashing
// ----------------------------------------------------------------------
// GET /slashing/validators/{validatorPubKey}/signing_info Get sign info of given validator
func getSigningInfo(t *testing.T, port string, validatorPubKey string) slashing.ValidatorSigningInfo {
	res, body := Request(t, port, "GET", fmt.Sprintf("/slashing/validators/%s/signing_info", validatorPubKey), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var signingInfo slashing.ValidatorSigningInfo
	err := cdc.UnmarshalJSON([]byte(body), &signingInfo)
	require.Nil(t, err)

	return signingInfo
}

// TODO: Test this functionality, it is not currently in any of the tests
// POST /slashing/validators/{validatorAddr}/unjail Unjail a jailed validator
func doUnjail(t *testing.T, port, seed, name, password string,
	valAddr sdk.ValAddress, fees sdk.Coins) (resultTx sdk.TxResponse) {
	chainID := viper.GetString(client.FlagChainID)
	baseReq := rest.NewBaseReq(name, password, "", chainID, "", "", 1, 1, fees, nil, false, false)

	ur := slashingrest.UnjailReq{
		BaseReq: baseReq,
	}
	req, err := cdc.MarshalJSON(ur)
	require.NoError(t, err)
	res, body := Request(t, port, "POST", fmt.Sprintf("/slashing/validators/%s/unjail", valAddr.String()), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results sdk.TxResponse
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

type unjailReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
}

// ICS24 - fee distribution

// POST /distribution/delegators/{delgatorAddr}/rewards Withdraw delegator rewards
func doWithdrawDelegatorAllRewards(t *testing.T, port, seed, name, password string,
	delegatorAddr sdk.AccAddress, fees sdk.Coins) (resultTx sdk.TxResponse) {
	// get the account to get the sequence
	acc := getAccount(t, port, delegatorAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	baseReq := rest.NewBaseReq(name, password, "", chainID, "", "", accnum, sequence, fees, nil, false, false)

	wr := struct {
		BaseReq rest.BaseReq `json:"base_req"`
	}{BaseReq: baseReq}

	req := cdc.MustMarshalJSON(wr)
	res, body := Request(t, port, "POST", fmt.Sprintf("/distribution/delegators/%s/rewards", delegatorAddr), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results sdk.TxResponse
	cdc.MustUnmarshalJSON([]byte(body), &results)

	return results
}

func mustParseDecCoins(dcstring string) sdk.DecCoins {
	dcoins, err := sdk.ParseDecCoins(dcstring)
	if err != nil {
		panic(err)
	}
	return dcoins
}
