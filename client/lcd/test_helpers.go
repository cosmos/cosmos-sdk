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

	cryptoKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/client/utils"
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

	txbuilder "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

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

// TODO: Make InitializeTestLCD safe to call in multiple tests at the same time
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
func getTransaction(t *testing.T, port string, hash string) tx.Info {
	var tx tx.Info
	res, body := Request(t, port, "GET", fmt.Sprintf("/txs/%s", hash), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &tx)
	require.NoError(t, err)
	return tx
}

// POST /txs broadcast txs

// GET /txs search transactions
func getTransactions(t *testing.T, port string, tags ...string) []tx.Info {
	var txs []tx.Info
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
func doKeysPost(t *testing.T, port, name, password, seed string) keys.KeyOutput {
	pk := postKeys{name, password, seed}
	req, err := cdc.MarshalJSON(pk)
	require.NoError(t, err)
	res, body := Request(t, port, "POST", "/keys", req)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err = cdc.UnmarshalJSON([]byte(body), &resp)
	require.Nil(t, err, body)
	return resp
}

type postKeys struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Seed     string `json:"seed"`
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

// POST /keys/{name}/recover Recover a account from a seed
func doRecoverKey(t *testing.T, port, recoverName, recoverPassword, seed string) {
	jsonStr := []byte(fmt.Sprintf(`{"password":"%s", "seed":"%s"}`, recoverPassword, seed))
	res, body := Request(t, port, "POST", fmt.Sprintf("/keys/%s/recover", recoverName), jsonStr)

	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resp keys.KeyOutput
	err := codec.Cdc.UnmarshalJSON([]byte(body), &resp)
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
	kr := updateKeyReq{oldPassword, newPassword}
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

type updateKeyReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// DELETE /keys/{name} Remove an account
func deleteKey(t *testing.T, port, name, password string) {
	dk := deleteKeyReq{password}
	req, err := cdc.MarshalJSON(dk)
	require.NoError(t, err)
	keyEndpoint := fmt.Sprintf("/keys/%s", name)
	res, body := Request(t, port, "DELETE", keyEndpoint, req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
}

type deleteKeyReq struct {
	Password string `json:"password"`
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
		Tx:               msg,
		LocalAccountName: name,
		Password:         password,
		ChainID:          chainID,
		AccountNumber:    accnum,
		Sequence:         sequence,
	}
	json, err := cdc.MarshalJSON(payload)
	require.Nil(t, err)
	res, body := Request(t, port, "POST", "/tx/sign", json)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &signedMsg))
	return signedMsg
}

// POST /tx/broadcast Send a signed Tx
func doBroadcast(t *testing.T, port string, msg auth.StdTx) ctypes.ResultBroadcastTxCommit {
	tx := broadcastReq{Tx: msg, Return: "block"}
	req, err := cdc.MarshalJSON(tx)
	require.Nil(t, err)
	res, body := Request(t, port, "POST", "/tx/broadcast", req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var resultTx ctypes.ResultBroadcastTxCommit
	require.Nil(t, cdc.UnmarshalJSON([]byte(body), &resultTx))
	return resultTx
}

type broadcastReq struct {
	Tx     auth.StdTx `json:"tx"`
	Return string     `json:"return"`
}

// GET /bank/balances/{address} Get the account balances

// POST /bank/accounts/{address}/transfers Send coins (build -> sign -> send)
func doTransfer(t *testing.T, port, seed, name, password string, addr sdk.AccAddress) (receiveAddr sdk.AccAddress, resultTx ctypes.ResultBroadcastTxCommit) {
	res, body, receiveAddr := doTransferWithGas(t, port, seed, name, password, addr, "", 0, false, false)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	err := cdc.UnmarshalJSON([]byte(body), &resultTx)
	require.Nil(t, err)

	return receiveAddr, resultTx
}

func doTransferWithGas(t *testing.T, port, seed, name, password string, addr sdk.AccAddress, gas string,
	gasAdjustment float64, simulate, generateOnly bool) (
	res *http.Response, body string, receiveAddr sdk.AccAddress) {

	// create receive address
	kb := client.MockKeyBase()
	receiveInfo, _, err := kb.CreateMnemonic("receive_address", cryptoKeys.English, gapp.DefaultKeyPass, cryptoKeys.SigningAlgo("secp256k1"))
	require.Nil(t, err)
	receiveAddr = sdk.AccAddress(receiveInfo.GetPubKey().Address())

	acc := getAccount(t, port, addr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)

	sr := sendReq{
		Amount: sdk.Coins{sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, 1)},
		BaseReq: utils.BaseReq{
			Name:          name,
			Password:      password,
			ChainID:       chainID,
			AccountNumber: accnum,
			Sequence:      sequence,
			Simulate:      simulate,
			GenerateOnly:  generateOnly,
		},
	}

	if len(gas) != 0 {
		sr.BaseReq.Gas = gas
	}

	if gasAdjustment > 0 {
		sr.BaseReq.GasAdjustment = fmt.Sprintf("%f", gasAdjustment)
	}

	req, err := cdc.MarshalJSON(sr)
	require.NoError(t, err)

	res, body = Request(t, port, "POST", fmt.Sprintf("/bank/accounts/%s/transfers", receiveAddr), req)
	return
}

type sendReq struct {
	Amount  sdk.Coins     `json:"amount"`
	BaseReq utils.BaseReq `json:"base_req"`
}

// ----------------------------------------------------------------------
// ICS 21 - Stake
// ----------------------------------------------------------------------

// POST /stake/delegators/{delegatorAddr}/delegations Submit delegation
func doDelegate(t *testing.T, port, name, password string,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	ed := msgDelegationsInput{
		BaseReq: utils.BaseReq{
			Name:          name,
			Password:      password,
			ChainID:       chainID,
			AccountNumber: accnum,
			Sequence:      sequence,
		},
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		Delegation:    sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, amount),
	}
	req, err := cdc.MarshalJSON(ed)
	require.NoError(t, err)
	res, body := Request(t, port, "POST", fmt.Sprintf("/stake/delegators/%s/delegations", delAddr.String()), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	var results ctypes.ResultBroadcastTxCommit
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)
	return results
}

type msgDelegationsInput struct {
	BaseReq       utils.BaseReq  `json:"base_req"`
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
	ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
	Delegation    sdk.Coin       `json:"delegation"`
}

// POST /stake/delegators/{delegatorAddr}/delegations Submit delegation
func doBeginUnbonding(t *testing.T, port, name, password string,
	delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()
	chainID := viper.GetString(client.FlagChainID)
	ed := msgBeginUnbondingInput{
		BaseReq: utils.BaseReq{
			Name:          name,
			Password:      password,
			ChainID:       chainID,
			AccountNumber: accnum,
			Sequence:      sequence,
		},
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		SharesAmount:  sdk.NewDec(amount),
	}
	req, err := cdc.MarshalJSON(ed)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/stake/delegators/%s/unbonding_delegations", delAddr), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

type msgBeginUnbondingInput struct {
	BaseReq       utils.BaseReq  `json:"base_req"`
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"` // in bech32
	ValidatorAddr sdk.ValAddress `json:"validator_addr"` // in bech32
	SharesAmount  sdk.Dec        `json:"shares"`
}

// POST /stake/delegators/{delegatorAddr}/delegations Submit delegation
func doBeginRedelegation(t *testing.T, port, name, password string,
	delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

	acc := getAccount(t, port, delAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)
	ed := msgBeginRedelegateInput{
		BaseReq: utils.BaseReq{
			Name:          name,
			Password:      password,
			ChainID:       chainID,
			AccountNumber: accnum,
			Sequence:      sequence,
		},
		DelegatorAddr:    delAddr,
		ValidatorSrcAddr: valSrcAddr,
		ValidatorDstAddr: valDstAddr,
		SharesAmount:     sdk.NewDec(amount),
	}
	req, err := cdc.MarshalJSON(ed)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/stake/delegators/%s/redelegations", delAddr), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

type msgBeginRedelegateInput struct {
	BaseReq          utils.BaseReq  `json:"base_req"`
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`     // in bech32
	ValidatorSrcAddr sdk.ValAddress `json:"validator_src_addr"` // in bech32
	ValidatorDstAddr sdk.ValAddress `json:"validator_dst_addr"` // in bech32
	SharesAmount     sdk.Dec        `json:"shares"`
}

// GET /stake/delegators/{delegatorAddr}/delegations Get all delegations from a delegator
func getDelegatorDelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) []stake.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/delegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var dels []stake.Delegation

	err := cdc.UnmarshalJSON([]byte(body), &dels)
	require.Nil(t, err)

	return dels
}

// GET /stake/delegators/{delegatorAddr}/unbonding_delegations Get all unbonding delegations from a delegator
func getDelegatorUnbondingDelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) []stake.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/unbonding_delegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var ubds []stake.UnbondingDelegation

	err := cdc.UnmarshalJSON([]byte(body), &ubds)
	require.Nil(t, err)

	return ubds
}

// GET /stake/delegators/{delegatorAddr}/redelegations Get all redelegations from a delegator
func getDelegatorRedelegations(t *testing.T, port string, delegatorAddr sdk.AccAddress) []stake.Redelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/redelegations", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var reds []stake.Redelegation

	err := cdc.UnmarshalJSON([]byte(body), &reds)
	require.Nil(t, err)

	return reds
}

// GET /stake/delegators/{delegatorAddr}/validators Query all validators that a delegator is bonded to
func getDelegatorValidators(t *testing.T, port string, delegatorAddr sdk.AccAddress) []stake.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/validators", delegatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bondedValidators []stake.Validator

	err := cdc.UnmarshalJSON([]byte(body), &bondedValidators)
	require.Nil(t, err)

	return bondedValidators
}

// GET /stake/delegators/{delegatorAddr}/validators/{validatorAddr} Query a validator that a delegator is bonded to
func getDelegatorValidator(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) stake.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/validators/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bondedValidator stake.Validator
	err := cdc.UnmarshalJSON([]byte(body), &bondedValidator)
	require.Nil(t, err)

	return bondedValidator
}

// GET /stake/delegators/{delegatorAddr}/txs Get all staking txs (i.e msgs) from a delegator
func getBondingTxs(t *testing.T, port string, delegatorAddr sdk.AccAddress, query string) []tx.Info {
	var res *http.Response
	var body string

	if len(query) > 0 {
		res, body = Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/txs?type=%s", delegatorAddr, query), nil)
	} else {
		res, body = Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/txs", delegatorAddr), nil)
	}
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var txs []tx.Info

	err := cdc.UnmarshalJSON([]byte(body), &txs)
	require.Nil(t, err)

	return txs
}

// GET /stake/delegators/{delegatorAddr}/delegations/{validatorAddr} Query the current delegation between a delegator and a validator
func getDelegation(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) stake.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/delegations/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var bond stake.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &bond)
	require.Nil(t, err)

	return bond
}

// GET /stake/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr} Query all unbonding delegations between a delegator and a validator
func getUndelegation(t *testing.T, port string, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) stake.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/delegators/%s/unbonding_delegations/%s", delegatorAddr, validatorAddr), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var unbond stake.UnbondingDelegation
	err := cdc.UnmarshalJSON([]byte(body), &unbond)
	require.Nil(t, err)

	return unbond
}

// GET /stake/validators Get all validator candidates
func getValidators(t *testing.T, port string) []stake.Validator {
	res, body := Request(t, port, "GET", "/stake/validators", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var validators []stake.Validator
	err := cdc.UnmarshalJSON([]byte(body), &validators)
	require.Nil(t, err)

	return validators
}

// GET /stake/validators/{validatorAddr} Query the information from a single validator
func getValidator(t *testing.T, port string, validatorAddr sdk.ValAddress) stake.Validator {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/validators/%s", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var validator stake.Validator
	err := cdc.UnmarshalJSON([]byte(body), &validator)
	require.Nil(t, err)

	return validator
}

// GET /stake/validators/{validatorAddr}/delegations Get all delegations from a validator
func getValidatorDelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []stake.Delegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/validators/%s/delegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var delegations []stake.Delegation
	err := cdc.UnmarshalJSON([]byte(body), &delegations)
	require.Nil(t, err)

	return delegations
}

// GET /stake/validators/{validatorAddr}/unbonding_delegations Get all unbonding delegations from a validator
func getValidatorUnbondingDelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []stake.UnbondingDelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/validators/%s/unbonding_delegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var ubds []stake.UnbondingDelegation
	err := cdc.UnmarshalJSON([]byte(body), &ubds)
	require.Nil(t, err)

	return ubds
}

// GET /stake/validators/{validatorAddr}/redelegations Get all outgoing redelegations from a validator
func getValidatorRedelegations(t *testing.T, port string, validatorAddr sdk.ValAddress) []stake.Redelegation {
	res, body := Request(t, port, "GET", fmt.Sprintf("/stake/validators/%s/redelegations", validatorAddr.String()), nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var reds []stake.Redelegation
	err := cdc.UnmarshalJSON([]byte(body), &reds)
	require.Nil(t, err)

	return reds
}

// GET /stake/pool Get the current state of the staking pool
func getStakePool(t *testing.T, port string) stake.Pool {
	res, body := Request(t, port, "GET", "/stake/pool", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)
	require.NotNil(t, body)
	var pool stake.Pool
	err := cdc.UnmarshalJSON([]byte(body), &pool)
	require.Nil(t, err)
	return pool
}

// GET /stake/parameters Get the current staking parameter values
func getStakeParams(t *testing.T, port string) stake.Params {
	res, body := Request(t, port, "GET", "/stake/parameters", nil)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var params stake.Params
	err := cdc.UnmarshalJSON([]byte(body), &params)
	require.Nil(t, err)
	return params
}

// ----------------------------------------------------------------------
// ICS 22 - Gov
// ----------------------------------------------------------------------
// POST /gov/proposals Submit a proposal
func doSubmitProposal(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)
	pr := postProposalReq{
		Title:          "Test",
		Description:    "test",
		ProposalType:   "Text",
		Proposer:       proposerAddr,
		InitialDeposit: sdk.Coins{sdk.NewCoin(stakeTypes.DefaultBondDenom, sdk.NewInt(amount))},
		BaseReq: utils.BaseReq{
			Name:          name,
			Password:      password,
			ChainID:       chainID,
			AccountNumber: accnum,
			Sequence:      sequence,
		},
	}

	req, err := cdc.MarshalJSON(pr)
	require.NoError(t, err)

	// submitproposal
	res, body := Request(t, port, "POST", "/gov/proposals", req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

type postProposalReq struct {
	BaseReq        utils.BaseReq  `json:"base_req"`
	Title          string         `json:"title"`           //  Title of the proposal
	Description    string         `json:"description"`     //  Description of the proposal
	ProposalType   string         `json:"proposal_type"`   //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Proposer       sdk.AccAddress `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins      `json:"initial_deposit"` // Coins to add to the proposal's deposit
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
func doDeposit(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, proposalID uint64, amount int64) (resultTx ctypes.ResultBroadcastTxCommit) {

	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)
	dr := depositReq{
		Depositor: proposerAddr,
		Amount:    sdk.Coins{sdk.NewCoin(stakeTypes.DefaultBondDenom, sdk.NewInt(amount))},
		BaseReq: utils.BaseReq{
			Name:          name,
			Password:      password,
			ChainID:       chainID,
			AccountNumber: accnum,
			Sequence:      sequence,
		},
	}

	req, err := cdc.MarshalJSON(dr)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/deposits", proposalID), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

type depositReq struct {
	BaseReq   utils.BaseReq  `json:"base_req"`
	Depositor sdk.AccAddress `json:"depositor"` // Address of the depositor
	Amount    sdk.Coins      `json:"amount"`    // Coins to add to the proposal's deposit
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
func doVote(t *testing.T, port, seed, name, password string, proposerAddr sdk.AccAddress, proposalID uint64) (resultTx ctypes.ResultBroadcastTxCommit) {
	// get the account to get the sequence
	acc := getAccount(t, port, proposerAddr)
	accnum := acc.GetAccountNumber()
	sequence := acc.GetSequence()

	chainID := viper.GetString(client.FlagChainID)
	vr := voteReq{
		Voter:  proposerAddr,
		Option: "Yes",
		BaseReq: utils.BaseReq{
			Name:          name,
			Password:      password,
			ChainID:       chainID,
			AccountNumber: accnum,
			Sequence:      sequence,
		},
	}

	req, err := cdc.MarshalJSON(vr)
	require.NoError(t, err)

	res, body := Request(t, port, "POST", fmt.Sprintf("/gov/proposals/%d/votes", proposalID), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results ctypes.ResultBroadcastTxCommit
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results
}

type voteReq struct {
	BaseReq utils.BaseReq  `json:"base_req"`
	Voter   sdk.AccAddress `json:"voter"`  //  address of the voter
	Option  string         `json:"option"` //  option from OptionSet chosen by the voter
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
	valAddr sdk.ValAddress) (resultTx ctypes.ResultBroadcastTxCommit) {
	chainID := viper.GetString(client.FlagChainID)

	ur := unjailReq{utils.BaseReq{
		Name:          name,
		Password:      password,
		ChainID:       chainID,
		AccountNumber: 1,
		Sequence:      1,
	}}
	req, err := cdc.MarshalJSON(ur)
	require.NoError(t, err)
	res, body := Request(t, port, "POST", fmt.Sprintf("/slashing/validators/%s/unjail", valAddr.String()), req)
	require.Equal(t, http.StatusOK, res.StatusCode, body)

	var results []ctypes.ResultBroadcastTxCommit
	err = cdc.UnmarshalJSON([]byte(body), &results)
	require.Nil(t, err)

	return results[0]
}

type unjailReq struct {
	BaseReq utils.BaseReq `json:"base_req"`
}
