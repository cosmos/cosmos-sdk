package clitest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmCfg "github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	keyFoo                    = "foo"
	keyBar                    = "bar"
	tokenFoo                  = "fooToken"
	bondDenom                 = stakeTypes.DefaultBondDenom
	defaultBondDenomAmount    = 150
	defaultAccountTokenAmount = 1000
)

func NewFixtures(t *testing.T) Fixtures {
	tmpDir := os.TempDir()
	gaiadHome := fmt.Sprintf("%s%s%s%s.test_gaiad", tmpDir, string(os.PathSeparator), t.Name(), string(os.PathSeparator))
	gaiacliHome := fmt.Sprintf("%s%s%s%s.test_gaiacli", tmpDir, string(os.PathSeparator), t.Name(), string(os.PathSeparator))
	rpcAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	p2pAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err)
	return Fixtures{
		T:           t,
		gaiadHome:   gaiadHome,
		gaiacliHome: gaiacliHome,
		rpcAddr:     rpcAddr,
		p2pAddr:     p2pAddr,
		port:        port,
	}
}

// Fixtures contains all the configuration parameters for a self contained test
type Fixtures struct {
	T *testing.T

	gaiadHome   string
	gaiacliHome string
	chainID     string
	port        string
	rpcAddr     string
	p2pAddr     string
}

// UnsafeResetAll clears directories for new usage
func (f Fixtures) ResetTest() {
	tests.ExecuteT(f.T, fmt.Sprintf("gaiad --home=%s unsafe-reset-all", f.gaiadHome), "")
	os.RemoveAll(filepath.Join(f.gaiadHome, "config", "gentx"))
	executeWrite(f.T, fmt.Sprintf("gaiacli keys delete --home=%s --force foo", f.gaiacliHome), app.DefaultKeyPass)
	executeWrite(f.T, fmt.Sprintf("gaiacli keys delete --home=%s --force bar", f.gaiacliHome), app.DefaultKeyPass)
}

// Init runs gaiad init and saves the chain-id on the struct
func (f Fixtures) Init() {
	f.ResetTest()
	_, stderr := tests.ExecuteT(f.T, fmt.Sprintf("gaiad init -o --moniker=foo --home=%s", f.gaiadHome), app.DefaultKeyPass)
	var chainID string
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(stderr), &initRes)
	require.NoError(f.T, err)

	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(f.T, err)

	f.chainID = chainID
	cfg := tmCfg.DefaultConfig()

	cfg.Consensus.TimeoutPrevote = time.Millisecond * 500
	cfg.Consensus.TimeoutPropose = time.Second * 1
	cfg.Consensus.TimeoutPrecommit = time.Millisecond * 500
	cfg.Consensus.TimeoutPrevoteDelta = time.Millisecond * 100
	cfg.Consensus.TimeoutPrecommitDelta = time.Millisecond * 100
	cfg.Consensus.TimeoutPrevoteDelta = time.Millisecond * 100
	cfg.Consensus.TimeoutCommit = time.Second * 1

	err = cfg.ValidateBasic()
	require.NoError(f.T, err)

	tmCfg.WriteConfigFile(filepath.Join(f.gaiadHome, "config", "config.toml"), cfg)
}

// AddKey adds a key of name to the key storage
func (f Fixtures) AddKey(name string) {
	executeWriteCheckErr(f.T, fmt.Sprintf("gaiacli keys add --home=%s foo", f.gaiacliHome), app.DefaultKeyPass)
}

// InitFixtures creats the common environment for most of the CLI tests
func InitFixtures(t *testing.T) Fixtures {
	// Ensure testing directories are clean and properly initialized
	f := NewFixtures(t)

	// Reset directories and set configuration
	f.Init()

	// Create new keys foo, bar to use in tests
	f.AddKey(keyFoo)
	f.AddKey(keyBar)
	t.Error("keys added")

	// Add foo account to genesis file with some coins in it for testing and finalize genesis
	f.ExecuteAddAccount(f.GetKeyAddress(keyFoo), tokenFoo)
	t.Error("account added to genesis")

	// generate a gentx to include bonding in genesis file
	f.GenTx(keyFoo)
	t.Error("gentx ran")

	// Collect all gentxs and finalize genesis
	f.CollectGenTxs()
	t.Error("collect gentx ran")
	return f
}

// UnmarsalStdTx unmarshals a standard transaction
func (f Fixtures) UnmarshalStdTx(s string) (stdTx auth.StdTx) {
	cdc := app.MakeCodec()
	require.Nil(f.T, cdc.UnmarshalJSON([]byte(s), &stdTx))
	return
}

func writeToNewTempFile(t *testing.T, s string) *os.File {
	fp, err := ioutil.TempFile(os.TempDir(), "cosmos_cli_test_")
	require.Nil(t, err)
	_, err = fp.WriteString(s)
	require.Nil(t, err)
	return fp
}

//___________________________________________________________________________________
// executors

func executeWriteCheckErr(t *testing.T, cmdStr string, writes ...string) {
	require.True(t, executeWrite(t, cmdStr, writes...))
}

func executeWrite(t *testing.T, cmdStr string, writes ...string) (exitSuccess bool) {
	exitSuccess, stdout, stderr := executeWriteRetStdStreams(t, cmdStr, writes...)
	t.Log(stdout)
	t.Log(stderr)
	return
}

func executeWriteRetStdStreams(t *testing.T, cmdStr string, writes ...string) (bool, string, string) {
	proc := tests.GoExecuteTWithStdout(t, cmdStr)

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

// ExecuteAddAccount adds an account to genesis.json
func (f Fixtures) ExecuteAddAccount(address sdk.AccAddress, tokenDenom string) {
	executeWriteCheckErr(f.T, fmt.Sprintf(
		"gaiad add-genesis-account %s %d%s,%d%s --home=%s",
		address, defaultBondDenomAmount, bondDenom,
		defaultAccountTokenAmount, tokenDenom, f.gaiadHome),
	)
}

// GenTx generates a gentx transaction
func (f Fixtures) GenTx(name string) {
	executeWriteCheckErr(f.T,
		fmt.Sprintf("gaiad gentx --name=foo --home=%s --home-client=%s", f.gaiadHome, f.gaiacliHome),
		app.DefaultKeyPass,
	)
}

// CollectGenTxs collects gentxs and finalizes genesis file
func (f Fixtures) CollectGenTxs() {
	executeWriteCheckErr(f.T, fmt.Sprintf("gaiad collect-gentxs --home=%s", f.gaiadHome), app.DefaultKeyPass)
}

//___________________________________________________________________________________
// txs

func (f Fixtures) executeGetTxs(t *testing.T, tags ...string) []tx.Info {
	if len(tags) < 1 {
		t.Error("Must pass tags into the transaction search function")
	}

	// Construct query
	tgs := ""
	for _, t := range tags {
		tgs += fmt.Sprintf("%s&", t)
	}
	tgs = strings.TrimSuffix(tgs, "&")

	// Run query
	cmdStr := fmt.Sprintf("gaiacli query txs --tags='%s' %v", tgs, f.GaiadFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var txs []tx.Info
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &txs)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return txs
}

//___________________________________________________________________________________
// stake

func (f Fixtures) executeGetValidator(t *testing.T, address sdk.ValAddress) stake.Validator {
	cmdStr := fmt.Sprintf("gaiacli query stake validator %s --output=json %v", address, f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var validator stake.Validator
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &validator)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return validator
}

func (f Fixtures) executeGetValidatorUnbondingDelegations(t *testing.T, address sdk.ValAddress) []stake.UnbondingDelegation {
	cmdStr := fmt.Sprintf("gaiacli query stake unbonding-delegations-from %s --output=json %v", address, f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var ubds []stake.UnbondingDelegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &ubds)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return ubds
}

func (f Fixtures) executeUnbondValidator(t *testing.T, from string, sharesAmount int, validator sdk.ValAddress) bool {
	unbondStr := fmt.Sprintf("gaiacli tx stake unbond %v", f.GaiaCliFlags())
	unbondStr += fmt.Sprintf(" --from=%s", from)
	unbondStr += fmt.Sprintf(" --validator=%s", validator)
	unbondStr += fmt.Sprintf(" --shares-amount=%d", sharesAmount)
	return executeWrite(t, unbondStr, app.DefaultKeyPass)
}

func executeGetValidatorRedelegations(t *testing.T, cmdStr string) []stake.Redelegation {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var reds []stake.Redelegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &reds)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return reds
}

func (f Fixtures) executeGetValidatorDelegations(t *testing.T, valAddr sdk.ValAddress) []stake.Delegation {
	cmdStr := fmt.Sprintf("gaiacli query stake delegations-to %s --output=json %v", valAddr, f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var delegations []stake.Delegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &delegations)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return delegations
}

func (f Fixtures) executeGetPool(t *testing.T) stake.Pool {
	cmdStr := fmt.Sprintf("gaiacli query stake pool --output=json %v", f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var pool stake.Pool
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &pool)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return pool
}

func (f Fixtures) executeGetStakeParams(t *testing.T) stake.Params {
	cmdStr := fmt.Sprintf("gaiacli query stake parameters --output=json %v", f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var params stake.Params
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &params)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return params
}

//___________________________________________________________________________________
// gov

func (f Fixtures) executeGetDepositParam(t *testing.T) gov.DepositParams {
	cmdStr := fmt.Sprintf("gaiacli query gov param deposit %v", f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var depositParam gov.DepositParams
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &depositParam)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return depositParam
}

func (f Fixtures) executeGetVotingParam(t *testing.T) gov.VotingParams {
	cmdStr := fmt.Sprintf("gaiacli query gov param deposit %v", f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var votingParam gov.VotingParams
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &votingParam)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return votingParam
}

func (f Fixtures) executeGetTallyingParam(t *testing.T) gov.TallyParams {
	cmdStr := fmt.Sprintf("gaiacli query gov param deposit %v", f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var tallyingParam gov.TallyParams
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &tallyingParam)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return tallyingParam
}

func (f Fixtures) executeGetProposal(t *testing.T, proposalID int) gov.Proposal {
	cmdStr := fmt.Sprintf("gaiacli query gov proposal --proposal-id %d --output=json %v", proposalID, f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var proposal gov.Proposal
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &proposal)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return proposal
}

func (f Fixtures) executeGetVote(t *testing.T, proposalID int, voter sdk.AccAddress) gov.Vote {
	cmdStr := fmt.Sprintf("gaiacli query gov vote --proposal-id %d --voter %s --output=json %v", proposalID, voter, f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var vote gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &vote)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return vote
}

func (f Fixtures) executeGetVotes(t *testing.T, proposalID int) []gov.Vote {
	cmdStr := fmt.Sprintf("gaiacli query gov votes --proposal-id %d --output=json %v", proposalID, f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var votes []gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &votes)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return votes
}

func (f Fixtures) executeGetDeposit(t *testing.T, proposalID int, depositer sdk.AccAddress) gov.Deposit {
	cmdStr := fmt.Sprintf("gaiacli query gov deposit --proposal-id %d --depositor %s --output=json %v", proposalID, depositer, f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var deposit gov.Deposit
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &deposit)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return deposit
}

func (f Fixtures) executeGetDeposits(t *testing.T, proposalID int) []gov.Deposit {
	cmdStr := fmt.Sprintf("gaiacli query gov deposits --proposal-id %d --output=json %v", proposalID, f.GaiaCliFlags())
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var deposits []gov.Deposit
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &deposits)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return deposits
}

// GaiadFlags returns the flags for working with gaiad
func (f Fixtures) GaiadFlags() string {
	return fmt.Sprintf("--home=%s --chain-id=%v", f.gaiadHome, f.chainID)
}

// GaiaCliFlags returns the flags for working with gaiacli
func (f Fixtures) GaiaCliFlags() string {
	return fmt.Sprintf("--home=%s --node=%v --chain-id=%v", f.gaiacliHome, f.rpcAddr, f.chainID)
}

// Cleanup dirs cleans up the directories used for testing
func (f Fixtures) CleanupDirs() {
	for _, d := range []string{f.gaiacliHome, f.gaiadHome} {
		os.RemoveAll(d)
	}
}

// StartGaiad starts a gaiad instance given the parameters from the fixtures instance
func (f Fixtures) StartGaiad(t *testing.T, flags ...string) *tests.Process {
	cmdStr := fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", f.gaiadHome, f.rpcAddr, f.p2pAddr)

	if len(flags) > 0 {
		for _, f := range flags {
			cmdStr += fmt.Sprintf(" %s", f)
		}
	}
	proc := tests.GoExecuteTWithStdout(t, cmdStr)
	tests.WaitForTMStart(f.port)
	tests.WaitForNextNBlocksTM(1, f.port)
	return proc
}

// GetKeyAddress gives back the sdk.AccAddress of a key from storage given a name
func (f Fixtures) GetKeyAddress(name string) sdk.AccAddress {
	cmdStr := fmt.Sprintf("gaiacli keys show %s --output=json --home=%s", name, f.gaiacliHome)
	out, _ := tests.ExecuteT(f.T, cmdStr, "")
	var ko keys.KeyOutput
	keys.UnmarshalJSON([]byte(out), &ko)
	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(f.T, err)
	return accAddr
}

func (f Fixtures) sendCoins(t *testing.T, amount, from string, to sdk.AccAddress, flags ...string) bool {
	cmdStr := fmt.Sprintf("gaiacli tx send --amount=%s --from=%s --to=%s %s", amount, from, to, f.GaiaCliFlags())
	if len(flags) > 0 {
		for _, f := range flags {
			cmdStr += fmt.Sprintf(" %s", f)
		}
	}
	return executeWrite(t, cmdStr, app.DefaultKeyPass)
}

func (f Fixtures) sendCoinsResponse(t *testing.T, amount, from string, to sdk.AccAddress, flags ...string) auth.StdTx {
	cmdStr := fmt.Sprintf("gaiacli tx send --json --amount=%s --from=%s --to=%s %s", amount, from, to, f.GaiaCliFlags())
	if len(flags) > 0 {
		for _, f := range flags {
			cmdStr += fmt.Sprintf(" %s", f)
		}
	}
	success, stdout, stderr := executeWriteRetStdStreams(t, cmdStr)
	require.True(f.T, success)
	require.Empty(f.T, stderr)
	stdTx := f.UnmarshalStdTx(stdout)
	return stdTx
}

// GetAccount returns the auth.BaseAccount associated with an addr
func (f Fixtures) GetAccount(addr sdk.AccAddress) auth.BaseAccount {
	// Construct query and execute
	cmdStr := fmt.Sprintf("gaiacli query account %s %v", addr, f.GaiaCliFlags())
	out, _ := tests.ExecuteT(f.T, cmdStr, "")

	// Unmarshal response
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(f.T, err, "out %v, err %v", out, err)

	// Pull out the account from the response
	value := initRes["value"]
	var acc auth.BaseAccount

	// Unmarshal the account
	cdc := codec.New()
	codec.RegisterCrypto(cdc)
	err = cdc.UnmarshalJSON(value, &acc)
	require.NoError(f.T, err, "value %v, err %v", string(value), err)

	return acc
}

func bondDenomTokens(i int) string {
	return fmt.Sprintf("%d%s", i, bondDenom)
}

func fooTokens(i int) string {
	return fmt.Sprintf("%d%s", i, tokenFoo)
}

func feeFlag(fee string) string {
	return fmt.Sprintf("--fee=%s", fee)
}

// DeliverTxResponse returns the response from deliver TX
type DeliverTxResponse struct {
	Response abci.ResponseDeliverTx
}
