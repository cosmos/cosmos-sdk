package clitest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tendermint/tendermint/types"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
)

const (
	denom    = stakeTypes.DefaultBondDenom
	keyFoo   = "foo"
	keyBar   = "bar"
	fooDenom = "footoken"
	feeDenom = "feetoken"
)

var startCoins = sdk.Coins{
	sdk.NewInt64Coin(feeDenom, 1000),
	sdk.NewInt64Coin(fooDenom, 1000),
	sdk.NewInt64Coin(denom, 150),
}

// Fixtures is used to setup the testing environment
type Fixtures struct {
	ChainID  string
	RPCAddr  string
	Port     string
	GDHome   string
	GCLIHome string
	P2PAddr  string
	T        *testing.T
}

func NewFixtures(t *testing.T) *Fixtures {
	tmpDir := os.TempDir()
	gaiadHome := fmt.Sprintf("%s%s%s%s.test_gaiad", tmpDir, string(os.PathSeparator), t.Name(), string(os.PathSeparator))
	gaiacliHome := fmt.Sprintf("%s%s%s%s.test_gaiacli", tmpDir, string(os.PathSeparator), t.Name(), string(os.PathSeparator))
	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	p2pAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)
	return &Fixtures{
		T:        t,
		GDHome:   gaiadHome,
		GCLIHome: gaiacliHome,
		RPCAddr:  servAddr,
		P2PAddr:  p2pAddr,
		Port:     port,
	}
}

// Cleanup is meant to be run at the end of a test to clean up an remaining test state
func (f *Fixtures) Cleanup() {
	cleanupDirs(f.GDHome, f.GCLIHome)
}

// Flags returns the flags necessary for making most CLI calls
func (f *Fixtures) Flags() string {
	return fmt.Sprintf("--home=%s --node=%s --chain-id=%s", f.GCLIHome, f.RPCAddr, f.ChainID)
}

func getTestingHomeDirs(name string) (string, string) {
	tmpDir := os.TempDir()
	gaiadHome := fmt.Sprintf("%s%s%s%s.test_gaiad", tmpDir, string(os.PathSeparator), name, string(os.PathSeparator))
	gaiacliHome := fmt.Sprintf("%s%s%s%s.test_gaiacli", tmpDir, string(os.PathSeparator), name, string(os.PathSeparator))
	return gaiadHome, gaiacliHome
}

// UnsafeResetAll is gaiad unsafe-reset-all
func (f *Fixtures) UnsafeResetAll(flags ...string) {
	cmd := fmt.Sprintf("gaiad --home=%s unsafe-reset-all", f.GDHome)
	executeWrite(f.T, addFlags(cmd, flags))
	os.RemoveAll(filepath.Join(f.GDHome, "config", "gentx"))
}

// KeysDelete is gaiacli keys delete
func (f *Fixtures) KeysDelete(name string, flags ...string) {
	cmd := fmt.Sprintf("gaiacli keys delete --home=%s %s", f.GCLIHome, name)
	executeWrite(f.T, addFlags(cmd, flags), app.DefaultKeyPass)
}

// KeysAdd is gaiacli keys add
func (f *Fixtures) KeysAdd(name string, flags ...string) {
	cmd := fmt.Sprintf("gaiacli keys add --home=%s %s", f.GCLIHome, name)
	executeWriteCheckErr(f.T, addFlags(cmd, flags), app.DefaultKeyPass)
}

// KeysShow is gaiacli keys show
func (f *Fixtures) KeysShow(name string, flags ...string) keys.KeyOutput {
	cmd := fmt.Sprintf("gaiacli keys show --home=%s %s", f.GCLIHome, name)
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var ko keys.KeyOutput
	err := keys.UnmarshalJSON([]byte(out), &ko)
	require.NoError(f.T, err)
	return ko
}

// KeyAddress returns the SDK account address from the key
func (f *Fixtures) KeyAddress(name string) sdk.AccAddress {
	ko := f.KeysShow(name)
	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(f.T, err)
	return accAddr
}

// CLIConfig is gaiacli config
func (f *Fixtures) CLIConfig(key, value string, flags ...string) {
	cmd := fmt.Sprintf("gaiacli config --home=%s %s %s", f.GCLIHome, key, value)
	executeWriteCheckErr(f.T, addFlags(cmd, flags))
}

// GDInit is gaiad init
// NOTE: GDInit sets the ChainID for the Fixtures instance
func (f *Fixtures) GDInit(moniker string, flags ...string) {
	cmd := fmt.Sprintf("gaiad init -o --moniker=%s --home=%s", moniker, f.GDHome)
	_, stderr := tests.ExecuteT(f.T, addFlags(cmd, flags), app.DefaultKeyPass)

	var chainID string
	var initRes map[string]json.RawMessage

	err := json.Unmarshal([]byte(stderr), &initRes)
	require.NoError(f.T, err)

	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(f.T, err)

	f.ChainID = chainID
}

func addFlags(cmd string, flags []string) string {
	for _, f := range flags {
		cmd += " " + f
	}
	return strings.TrimSpace(cmd)
}

// AddGenesisAccount is gaiad add-genesis-account
func (f *Fixtures) AddGenesisAccount(address sdk.AccAddress, coins sdk.Coins, flags ...string) {
	cmd := fmt.Sprintf("gaiad add-genesis-account %s %s --home=%s", address, coins, f.GDHome)
	executeWriteCheckErr(f.T, addFlags(cmd, flags))
}

// GenTx is gaiad gentx
func (f *Fixtures) GenTx(name string, flags ...string) {
	cmd := fmt.Sprintf("gaiad gentx --name=%s --home=%s --home-client=%s", name, f.GDHome, f.GCLIHome)
	executeWriteCheckErr(f.T, addFlags(cmd, flags), app.DefaultKeyPass)
}

// CollectGenTxs is gaiad collect-gentxs
func (f *Fixtures) CollectGenTxs(flags ...string) {
	cmd := fmt.Sprintf("gaiad collect-gentxs --home=%s", f.GDHome)
	executeWriteCheckErr(f.T, addFlags(cmd, flags), app.DefaultKeyPass)
}

// GDStart runs gaiad start with the appropriate flags and returns a process
func (f *Fixtures) GDStart(flags ...string) *tests.Process {
	cmd := fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", f.GDHome, f.RPCAddr, f.P2PAddr)
	proc := tests.GoExecuteTWithStdout(f.T, addFlags(cmd, flags))
	tests.WaitForTMStart(f.Port)
	tests.WaitForNextNBlocksTM(1, f.Port)
	return proc
}

func initializeFixtures(t *testing.T) (f *Fixtures) {
	f = NewFixtures(t)

	// Reset test state
	f.UnsafeResetAll()

	// Ensure keystore has foo and bar keys
	f.KeysDelete(keyFoo)
	f.KeysDelete(keyBar)
	f.KeysAdd(keyFoo)
	f.KeysAdd(keyBar)

	// Ensure that CLI output is in JSON format
	f.CLIConfig("output", "json")

	// NOTE: GDInit sets the ChainID
	f.GDInit(keyFoo)

	// Start an account with tokens
	f.AddGenesisAccount(f.KeyAddress(keyFoo), startCoins)
	f.GenTx(keyFoo)
	f.CollectGenTxs()
	return
}

// TxSend is gaiacli tx send
func (f *Fixtures) TxSend(from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) bool {
	cmd := fmt.Sprintf("gaiacli tx send %v --amount=%s --to=%s --from=%s", f.Flags(), amount, to, from)
	return executeWrite(f.T, addFlags(cmd, flags), app.DefaultKeyPass)
}

type sendResponse struct {
	Height   int64
	TxHash   string
	Response abci.ResponseDeliverTx
}

// TxSendWResponse is gaiacli tx send an also returns the complete response
func (f *Fixtures) TxSendWResponse(from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) sendResponse {
	cmd := fmt.Sprintf("gaiacli tx send --json %v --amount=%s --to=%s --from=%s", f.Flags(), amount, to, from)
	success, stdout, _ := executeWriteRetStdStreams(f.T, addFlags(cmd, flags), app.DefaultKeyPass)
	require.True(f.T, success)
	cdc := app.MakeCodec()
	var jsonOutput sendResponse
	err := cdc.UnmarshalJSON([]byte(stdout), &jsonOutput)
	require.Nil(f.T, err)
	return jsonOutput
}

// TxStakeCreateValidator is gaiacli tx stake create-validator
func (f *Fixtures) TxStakeCreateValidator(from, consPubKey string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("gaiacli tx stake create-validator %v --from=%s --pubkey=%s", f.Flags(), from, consPubKey)
	cmd += fmt.Sprintf(" --amount=%v --moniker=%v --commission-rate=%v", amount, from, "0.05")
	cmd += fmt.Sprintf(" --commission-max-rate=%v --commission-max-change-rate=%v", "0.20", "0.10")
	return executeWriteRetStdStreams(f.T, addFlags(cmd, flags), app.DefaultKeyPass)
}

// TxStakeUnbond is gaiacli tx stake unbond
func (f *Fixtures) TxStakeUnbond(from, shares string, validator sdk.ValAddress, flags ...string) bool {
	cmd := fmt.Sprintf("gaiacli tx stake unbond %v --from=%s --validator=%s --shares-amount=%v", f.Flags(), from, validator, shares)
	return executeWrite(f.T, addFlags(cmd, flags), app.DefaultKeyPass)
}

// NEW
// -------------------------------------------
// OLD

func marshalStdTx(t *testing.T, stdTx auth.StdTx) []byte {
	cdc := app.MakeCodec()
	bz, err := cdc.MarshalBinaryBare(stdTx)
	require.NoError(t, err)
	return bz
}

func unmarshalStdTx(t *testing.T, s string) (stdTx auth.StdTx) {
	cdc := app.MakeCodec()
	require.Nil(t, cdc.UnmarshalJSON([]byte(s), &stdTx))
	return
}

func writeToNewTempFile(t *testing.T, s string) *os.File {
	fp, err := ioutil.TempFile(os.TempDir(), "cosmos_cli_test_")
	require.Nil(t, err)
	_, err = fp.WriteString(s)
	require.Nil(t, err)
	return fp
}

func readGenesisFile(t *testing.T, genFile string) types.GenesisDoc {
	var genDoc types.GenesisDoc
	fp, err := os.Open(genFile)
	require.NoError(t, err)
	fileContents, err := ioutil.ReadAll(fp)
	require.NoError(t, err)
	defer fp.Close()
	err = codec.Cdc.UnmarshalJSON(fileContents, &genDoc)
	require.NoError(t, err)
	return genDoc
}

//___________________________________________________________________________________
// executors

func executeWriteCheckErr(t *testing.T, cmdStr string, writes ...string) {
	require.True(t, executeWrite(t, cmdStr, writes...))
}

func executeWrite(t *testing.T, cmdStr string, writes ...string) (exitSuccess bool) {
	exitSuccess, _, _ = executeWriteRetStdStreams(t, cmdStr, writes...)
	return
}

func executeWriteRetStdStreams(t *testing.T, cmdStr string, writes ...string) (bool, string, string) {
	proc := tests.GoExecuteT(t, cmdStr)

	for _, write := range writes {
		_, err := proc.StdinPipe.Write([]byte(write + "\n"))
		require.NoError(t, err)
	}
	stdout, stderr, err := proc.ReadAll()
	if err != nil {
		fmt.Println("Err on proc.ReadAll()", err, cmdStr)
	}
	// Log output.
	if len(stdout) > 0 {
		t.Log("Stdout:", cmn.Green(string(stdout)))
	}
	if len(stderr) > 0 {
		t.Log("Stderr:", cmn.Red(string(stderr)))
	}

	proc.Wait()
	return proc.ExitState.Success(), string(stdout), string(stderr)
}

func executeInit(t *testing.T, cmdStr string) (chainID string) {
	_, stderr := tests.ExecuteT(t, cmdStr, app.DefaultKeyPass)

	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(stderr), &initRes)
	require.NoError(t, err)

	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(t, err)

	return
}

func executeGetAddrPK(t *testing.T, cmdStr string) (sdk.AccAddress, crypto.PubKey) {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var ko keys.KeyOutput
	keys.UnmarshalJSON([]byte(out), &ko)

	pk, err := sdk.GetAccPubKeyBech32(ko.PubKey)
	require.NoError(t, err)

	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(t, err)

	return accAddr, pk
}

// QueryAccount is gaiacli query account
func (f *Fixtures) QueryAccount(address sdk.AccAddress, flags ...string) auth.BaseAccount {
	cmd := fmt.Sprintf("gaiacli query account %s %v", address, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(f.T, err, "out %v, err %v", out, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	err = cdc.UnmarshalJSON(value, &acc)
	require.NoError(f.T, err, "value %v, err %v", string(value), err)
	return acc
}

// TODO: Remove
func executeGetAccount(t *testing.T, cmdStr string) auth.BaseAccount {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(t, err, "out %v, err %v", out, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	err = cdc.UnmarshalJSON(value, &acc)
	require.NoError(t, err, "value %v, err %v", string(value), err)
	return acc
}

//___________________________________________________________________________________
// txs

func executeGetTxs(t *testing.T, cmdStr string) []tx.Info {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var txs []tx.Info
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &txs)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return txs
}

//___________________________________________________________________________________
// stake

// QueryStakeValidator is gaiacli query stake validator
func (f *Fixtures) QueryStakeValidator(valAddr sdk.ValAddress, flags ...string) stake.Validator {
	cmd := fmt.Sprintf("gaiacli query stake validator %s %v", valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var validator stake.Validator
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &validator)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return validator
}

// QueryStakeUnbondingDelegationsFrom is gaiacli query stake unbonding-delegations-from
func (f *Fixtures) QueryStakeUnbondingDelegationsFrom(valAddr sdk.ValAddress, flags ...string) []stake.UnbondingDelegation {
	cmd := fmt.Sprintf("gaiacli query stake unbonding-delegations-from %s %v", valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var ubds []stake.UnbondingDelegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &ubds)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return ubds
}

func executeGetValidatorRedelegations(t *testing.T, cmdStr string) []stake.Redelegation {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var reds []stake.Redelegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &reds)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return reds
}

// QueryStakeDelegationsTo is gaiacli query stake delegations-to
func (f *Fixtures) QueryStakeDelegationsTo(valAddr sdk.ValAddress, flags ...string) []stake.Delegation {
	cmd := fmt.Sprintf("gaiacli query stake delegations-to %s %v", valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var delegations []stake.Delegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &delegations)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return delegations
}

// QueryStakePool is gaiacli query stake pool
func (f *Fixtures) QueryStakePool(flags ...string) stake.Pool {
	cmd := fmt.Sprintf("gaiacli query stake pool %v", f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var pool stake.Pool
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &pool)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return pool
}

// QueryStakeParameters is gaiacli query stake parameters
func (f *Fixtures) QueryStakeParameters(flags ...string) stake.Params {
	cmd := fmt.Sprintf("gaiacli query stake parameters %v", f.Flags())
	out, _ := tests.ExecuteT(f.T, addFlags(cmd, flags), "")
	var params stake.Params
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &params)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return params
}

//___________________________________________________________________________________
// gov

func executeGetDepositParam(t *testing.T, cmdStr string) gov.DepositParams {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var depositParam gov.DepositParams
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &depositParam)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return depositParam
}

func executeGetVotingParam(t *testing.T, cmdStr string) gov.VotingParams {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var votingParam gov.VotingParams
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &votingParam)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return votingParam
}

func executeGetTallyingParam(t *testing.T, cmdStr string) gov.TallyParams {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var tallyingParam gov.TallyParams
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &tallyingParam)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return tallyingParam
}

func executeGetProposal(t *testing.T, cmdStr string) gov.Proposal {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var proposal gov.Proposal
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &proposal)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return proposal
}

func executeGetVote(t *testing.T, cmdStr string) gov.Vote {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var vote gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &vote)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return vote
}

func executeGetVotes(t *testing.T, cmdStr string) []gov.Vote {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var votes []gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &votes)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return votes
}

func executeGetDeposit(t *testing.T, cmdStr string) gov.Deposit {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var deposit gov.Deposit
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &deposit)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return deposit
}

func executeGetDeposits(t *testing.T, cmdStr string) []gov.Deposit {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var deposits []gov.Deposit
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &deposits)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return deposits
}

func cleanupDirs(dirs ...string) {
	for _, d := range dirs {
		os.RemoveAll(d)
	}
}
