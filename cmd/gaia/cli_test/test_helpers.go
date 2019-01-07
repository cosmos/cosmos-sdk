package clitest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/tendermint/tendermint/types"

	"github.com/stretchr/testify/require"

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

//___________________________________________________________________________________
// helper methods

func getTestingHomeDirs(name string) (string, string) {
	tmpDir := os.TempDir()
	gaiadHome := fmt.Sprintf("%s%s%s%s.test_gaiad", tmpDir, string(os.PathSeparator), name, string(os.PathSeparator))
	gaiacliHome := fmt.Sprintf("%s%s%s%s.test_gaiacli", tmpDir, string(os.PathSeparator), name, string(os.PathSeparator))
	return gaiadHome, gaiacliHome
}

func initializeFixtures(t *testing.T) (chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr string) {
	gaiadHome, gaiacliHome = getTestingHomeDirs(t.Name())
	tests.ExecuteT(t, fmt.Sprintf("gaiad --home=%s unsafe-reset-all", gaiadHome), "")
	os.RemoveAll(filepath.Join(gaiadHome, "config", "gentx"))
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s foo", gaiacliHome), app.DefaultKeyPass)
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s bar", gaiacliHome), app.DefaultKeyPass)
	executeWriteCheckErr(t, fmt.Sprintf("gaiacli keys add --home=%s foo", gaiacliHome), app.DefaultKeyPass)
	executeWriteCheckErr(t, fmt.Sprintf("gaiacli keys add --home=%s bar", gaiacliHome), app.DefaultKeyPass)
	executeWriteCheckErr(t, fmt.Sprintf("gaiacli config --home=%s output json", gaiacliHome))
	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))
	chainID = executeInit(t, fmt.Sprintf("gaiad init -o --moniker=foo --home=%s", gaiadHome))

	executeWriteCheckErr(t, fmt.Sprintf(
		"gaiad add-genesis-account %s 150%s,1000footoken --home=%s", fooAddr, stakeTypes.DefaultBondDenom, gaiadHome))
	executeWrite(t, fmt.Sprintf("cat %s%sconfig%sgenesis.json", gaiadHome, string(os.PathSeparator), string(os.PathSeparator)))
	executeWriteCheckErr(t, fmt.Sprintf(
		"gaiad gentx --name=foo --home=%s --home-client=%s", gaiadHome, gaiacliHome), app.DefaultKeyPass)
	executeWriteCheckErr(t, fmt.Sprintf("gaiad collect-gentxs --home=%s", gaiadHome), app.DefaultKeyPass)
	// get a free port, also setup some common flags
	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	p2pAddr, _, err = server.FreeTCPAddr()
	require.NoError(t, err)
	return
}

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

func executeGetValidator(t *testing.T, cmdStr string) stake.Validator {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var validator stake.Validator
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &validator)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return validator
}

func executeGetValidatorUnbondingDelegations(t *testing.T, cmdStr string) []stake.UnbondingDelegation {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var ubds []stake.UnbondingDelegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &ubds)
	require.NoError(t, err, "out %v\n, err %v", out, err)
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

func executeGetValidatorDelegations(t *testing.T, cmdStr string) []stake.Delegation {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var delegations []stake.Delegation
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &delegations)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return delegations
}

func executeGetPool(t *testing.T, cmdStr string) stake.Pool {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var pool stake.Pool
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &pool)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return pool
}

func executeGetParams(t *testing.T, cmdStr string) stake.Params {
	out, _ := tests.ExecuteT(t, cmdStr, "")
	var params stake.Params
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &params)
	require.NoError(t, err, "out %v\n, err %v", out, err)
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
