package clitest

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

var (
	pass        = "1234567890"
	gaiadHome   = ""
	gaiacliHome = ""
)

func init() {
	gaiadHome, gaiacliHome = getTestingHomeDirs()
}

func TestGaiaCLISend(t *testing.T) {
	tests.ExecuteT(t, fmt.Sprintf("gaiad --home=%s unsafe_reset_all", gaiadHome))
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s foo", gaiacliHome), pass)
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s bar", gaiacliHome), pass)
	chainID := executeInit(t, fmt.Sprintf("gaiad init -o --name=foo --home=%s --home-client=%s", gaiadHome, gaiacliHome))
	executeWrite(t, fmt.Sprintf("gaiacli keys add --home=%s bar", gaiacliHome), pass)

	// get a free port, also setup some common flags
	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v", gaiadHome, servAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextHeightTM(port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --output=json --home=%s", gaiacliHome))
	fooCech := sdk.MustBech32ifyAcc(fooAddr)
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --output=json --home=%s", gaiacliHome))
	barCech := sdk.MustBech32ifyAcc(barAddr)

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooCech, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf("steak").Int64())

	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%v --from=foo", flags, barCech), pass)
	tests.WaitForNextHeightTM(port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barCech, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooCech, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf("steak").Int64())

	// test autosequencing
	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%v --from=foo", flags, barCech), pass)
	tests.WaitForNextHeightTM(port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barCech, flags))
	require.Equal(t, int64(20), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooCech, flags))
	require.Equal(t, int64(30), fooAcc.GetCoins().AmountOf("steak").Int64())

	// test memo
	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%v --from=foo --memo 'testmemo'", flags, barCech), pass)
	tests.WaitForNextHeightTM(port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barCech, flags))
	require.Equal(t, int64(30), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooCech, flags))
	require.Equal(t, int64(20), fooAcc.GetCoins().AmountOf("steak").Int64())
}

func TestGaiaCLICreateValidator(t *testing.T) {
	tests.ExecuteT(t, fmt.Sprintf("gaiad --home=%s unsafe_reset_all", gaiadHome))
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s foo", gaiacliHome), pass)
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s bar", gaiacliHome), pass)
	chainID := executeInit(t, fmt.Sprintf("gaiad init -o --name=foo --home=%s --home-client=%s", gaiadHome, gaiacliHome))
	executeWrite(t, fmt.Sprintf("gaiacli keys add --home=%s bar", gaiacliHome), pass)

	// get a free port, also setup some common flags
	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v", gaiadHome, servAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextHeightTM(port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --output=json --home=%s", gaiacliHome))
	fooCech := sdk.MustBech32ifyAcc(fooAddr)
	barAddr, barPubKey := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --output=json --home=%s", gaiacliHome))
	barCech := sdk.MustBech32ifyAcc(barAddr)
	barCeshPubKey := sdk.MustBech32ifyValPub(barPubKey)

	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%v --from=foo", flags, barCech), pass)
	tests.WaitForNextHeightTM(port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barCech, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooCech, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf("steak").Int64())

	// create validator
	cvStr := fmt.Sprintf("gaiacli stake create-validator %v", flags)
	cvStr += fmt.Sprintf(" --from=%v", "bar")
	cvStr += fmt.Sprintf(" --address-validator=%v", barCech)
	cvStr += fmt.Sprintf(" --pubkey=%v", barCeshPubKey)
	cvStr += fmt.Sprintf(" --amount=%v", "2steak")
	cvStr += fmt.Sprintf(" --moniker=%v", "bar-vally")

	executeWrite(t, cvStr, pass)
	tests.WaitForNextHeightTM(port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barCech, flags))
	require.Equal(t, int64(8), barAcc.GetCoins().AmountOf("steak").Int64(), "%v", barAcc)

	validator := executeGetValidator(t, fmt.Sprintf("gaiacli stake validator %v --output=json %v", barCech, flags))
	require.Equal(t, validator.Owner, barAddr)
	require.Equal(t, "2/1", validator.PoolShares.Amount.String())

	// unbond a single share
	unbondStr := fmt.Sprintf("gaiacli stake unbond begin %v", flags)
	unbondStr += fmt.Sprintf(" --from=%v", "bar")
	unbondStr += fmt.Sprintf(" --address-validator=%v", barCech)
	unbondStr += fmt.Sprintf(" --address-delegator=%v", barCech)
	unbondStr += fmt.Sprintf(" --shares-amount=%v", "1")

	success := executeWrite(t, unbondStr, pass)
	require.True(t, success)
	tests.WaitForNextHeightTM(port)

	/* // this won't be what we expect because we've only started unbonding, haven't completed
	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barCech, flags))
	require.Equal(t, int64(9), barAcc.GetCoins().AmountOf("steak").Int64(), "%v", barAcc)
	*/
	validator = executeGetValidator(t, fmt.Sprintf("gaiacli stake validator %v --output=json %v", barCech, flags))
	require.Equal(t, "1/1", validator.PoolShares.Amount.String())
}

func TestGaiaCLISubmitProposal(t *testing.T) {
	tests.ExecuteT(t, fmt.Sprintf("gaiad --home=%s unsafe_reset_all", gaiadHome))
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s foo", gaiacliHome), pass)
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s bar", gaiacliHome), pass)
	chainID := executeInit(t, fmt.Sprintf("gaiad init -o --name=foo --home=%s --home-client=%s", gaiadHome, gaiacliHome))
	executeWrite(t, fmt.Sprintf("gaiacli keys add --home=%s bar", gaiacliHome), pass)

	// get a free port, also setup some common flags
	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v", gaiadHome, servAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextHeightTM(port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --output=json --home=%s", gaiacliHome))
	fooCech := sdk.MustBech32ifyAcc(fooAddr)

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooCech, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf("steak").Int64())

	executeWrite(t, fmt.Sprintf("gaiacli gov submitproposal %v --proposer=%v --deposit=5steak --type=Text --title=Test --description=test --from=foo", flags, fooCech), pass)
	tests.WaitForNextHeightTM(port)

	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooCech, flags))
	require.Equal(t, int64(45), fooAcc.GetCoins().AmountOf("steak").Int64())

	proposal1 := executeGetProposal(t, fmt.Sprintf("gaiacli gov query-proposal --proposalID=1 --output=json %v", flags))
	require.Equal(t, int64(1), proposal1.ProposalID)
	require.Equal(t, gov.StatusToString(gov.StatusDepositPeriod), proposal1.Status)

	executeWrite(t, fmt.Sprintf("gaiacli gov deposit %v --depositer=%v --deposit=10steak --proposalID=1 --from=foo", flags, fooCech), pass)
	tests.WaitForNextHeightTM(port)

	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooCech, flags))
	require.Equal(t, int64(35), fooAcc.GetCoins().AmountOf("steak").Int64())
	proposal1 = executeGetProposal(t, fmt.Sprintf("gaiacli gov query-proposal --proposalID=1 --output=json %v", flags))
	require.Equal(t, int64(1), proposal1.ProposalID)
	require.Equal(t, gov.StatusToString(gov.StatusVotingPeriod), proposal1.Status)

	executeWrite(t, fmt.Sprintf("gaiacli gov vote %v --proposalID=1 --voter=%v --option=Yes --from=foo", flags, fooCech), pass)
	tests.WaitForNextHeightTM(port)

	vote := executeGetVote(t, fmt.Sprintf("gaiacli gov query-vote  --proposalID=1 --voter=%v --output=json %v", fooCech, flags))
	require.Equal(t, int64(1), vote.ProposalID)
	require.Equal(t, gov.VoteOptionToString(gov.OptionYes), vote.Option)
}

//___________________________________________________________________________________
// helper methods

func getTestingHomeDirs() (string, string) {
	tmpDir := os.TempDir()
	gaiadHome := fmt.Sprintf("%s%s.test_gaiad", tmpDir, string(os.PathSeparator))
	gaiacliHome := fmt.Sprintf("%s%s.test_gaiacli", tmpDir, string(os.PathSeparator))
	return gaiadHome, gaiacliHome
}

//___________________________________________________________________________________
// executors

func executeWrite(t *testing.T, cmdStr string, writes ...string) bool {
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
	return proc.ExitState.Success()
	//	bz := proc.StdoutBuffer.Bytes()
	//	fmt.Println("EXEC WRITE", string(bz))
}

func executeInit(t *testing.T, cmdStr string) (chainID string) {
	out := tests.ExecuteT(t, cmdStr)

	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(t, err)

	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(t, err)

	return
}

func executeGetAddrPK(t *testing.T, cmdStr string) (sdk.Address, crypto.PubKey) {
	out := tests.ExecuteT(t, cmdStr)
	var ko keys.KeyOutput
	keys.UnmarshalJSON([]byte(out), &ko)

	address, err := sdk.GetAccAddressBech32(ko.Address)
	require.NoError(t, err)

	pk, err := sdk.GetAccPubKeyBech32(ko.PubKey)
	require.NoError(t, err)

	return address, pk
}

func executeGetAccount(t *testing.T, cmdStr string) auth.BaseAccount {
	out := tests.ExecuteT(t, cmdStr)
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(t, err, "out %v, err %v", out, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	cdc := wire.NewCodec()
	wire.RegisterCrypto(cdc)
	err = cdc.UnmarshalJSON(value, &acc)
	require.NoError(t, err, "value %v, err %v", string(value), err)
	return acc
}

func executeGetValidator(t *testing.T, cmdStr string) stake.Validator {
	out := tests.ExecuteT(t, cmdStr)
	var validator stake.Validator
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &validator)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return validator
}

func executeGetProposal(t *testing.T, cmdStr string) gov.ProposalRest {
	out := tests.ExecuteT(t, cmdStr)
	var proposal gov.ProposalRest
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &proposal)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return proposal
}

func executeGetVote(t *testing.T, cmdStr string) gov.VoteRest {
	out := tests.ExecuteT(t, cmdStr)
	var vote gov.VoteRest
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &vote)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return vote
}
