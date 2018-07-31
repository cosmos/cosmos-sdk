// +build cli_test

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
	tests.WaitForNextNBlocksTM(2, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --output=json --home=%s", gaiacliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --output=json --home=%s", gaiacliHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf("steak").Int64())

	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%s --from=foo", flags, barAddr), pass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf("steak").Int64())

	// test autosequencing
	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%s --from=foo", flags, barAddr), pass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(20), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(30), fooAcc.GetCoins().AmountOf("steak").Int64())

	// test memo
	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%s --from=foo --memo 'testmemo'", flags, barAddr), pass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(30), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
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
	tests.WaitForNextNBlocksTM(2, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --output=json --home=%s", gaiacliHome))
	barAddr, barPubKey := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --output=json --home=%s", gaiacliHome))
	barCeshPubKey := sdk.MustBech32ifyValPub(barPubKey)

	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%s --from=foo", flags, barAddr), pass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf("steak").Int64())

	// create validator
	cvStr := fmt.Sprintf("gaiacli stake create-validator %v", flags)
	cvStr += fmt.Sprintf(" --from=%s", "bar")
	cvStr += fmt.Sprintf(" --pubkey=%s", barCeshPubKey)
	cvStr += fmt.Sprintf(" --amount=%v", "2steak")
	cvStr += fmt.Sprintf(" --moniker=%v", "bar-vally")

	executeWrite(t, cvStr, pass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(8), barAcc.GetCoins().AmountOf("steak").Int64(), "%v", barAcc)

	validator := executeGetValidator(t, fmt.Sprintf("gaiacli stake validator %s --output=json %v", barAddr, flags))
	require.Equal(t, validator.Owner, barAddr)
	require.True(sdk.RatEq(t, sdk.NewRat(2), validator.Tokens))

	// unbond a single share
	unbondStr := fmt.Sprintf("gaiacli stake unbond begin %v", flags)
	unbondStr += fmt.Sprintf(" --from=%s", "bar")
	unbondStr += fmt.Sprintf(" --address-validator=%s", barAddr)
	unbondStr += fmt.Sprintf(" --shares-amount=%v", "1")

	success := executeWrite(t, unbondStr, pass)
	require.True(t, success)
	tests.WaitForNextNBlocksTM(2, port)

	/* // this won't be what we expect because we've only started unbonding, haven't completed
	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barCech, flags))
	require.Equal(t, int64(9), barAcc.GetCoins().AmountOf("steak").Int64(), "%v", barAcc)
	*/
	validator = executeGetValidator(t, fmt.Sprintf("gaiacli stake validator %s --output=json %v", barAddr, flags))
	require.Equal(t, "1/1", validator.Tokens.String())
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
	tests.WaitForNextNBlocksTM(2, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --output=json --home=%s", gaiacliHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf("steak").Int64())

	proposalsQuery := tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals %v", flags))
	require.Equal(t, "No matching proposals found", proposalsQuery)

	// submit a test proposal
	spStr := fmt.Sprintf("gaiacli gov submit-proposal %v", flags)
	spStr += fmt.Sprintf(" --from=%s", "foo")
	spStr += fmt.Sprintf(" --deposit=%s", "5steak")
	spStr += fmt.Sprintf(" --type=%s", "Text")
	spStr += fmt.Sprintf(" --title=%s", "Test")
	spStr += fmt.Sprintf(" --description=%s", "test")

	executeWrite(t, spStr, pass)
	tests.WaitForNextNBlocksTM(2, port)

	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(45), fooAcc.GetCoins().AmountOf("steak").Int64())

	proposal1 := executeGetProposal(t, fmt.Sprintf("gaiacli gov query-proposal --proposal-id=1 --output=json %v", flags))
	require.Equal(t, int64(1), proposal1.GetProposalID())
	require.Equal(t, gov.StatusDepositPeriod, proposal1.GetStatus())

	proposalsQuery = tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals %v", flags))
	require.Equal(t, "  1 - Test", proposalsQuery)

	depositStr := fmt.Sprintf("gaiacli gov deposit %v", flags)
	depositStr += fmt.Sprintf(" --from=%s", "foo")
	depositStr += fmt.Sprintf(" --deposit=%s", "10steak")
	depositStr += fmt.Sprintf(" --proposal-id=%s", "1")

	executeWrite(t, depositStr, pass)
	tests.WaitForNextNBlocksTM(2, port)

	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(35), fooAcc.GetCoins().AmountOf("steak").Int64())
	proposal1 = executeGetProposal(t, fmt.Sprintf("gaiacli gov query-proposal --proposal-id=1 --output=json %v", flags))
	require.Equal(t, int64(1), proposal1.GetProposalID())
	require.Equal(t, gov.StatusVotingPeriod, proposal1.GetStatus())

	voteStr := fmt.Sprintf("gaiacli gov vote %v", flags)
	voteStr += fmt.Sprintf(" --from=%s", "foo")
	voteStr += fmt.Sprintf(" --proposal-id=%s", "1")
	voteStr += fmt.Sprintf(" --option=%s", "Yes")

	executeWrite(t, voteStr, pass)
	tests.WaitForNextNBlocksTM(2, port)

	vote := executeGetVote(t, fmt.Sprintf("gaiacli gov query-vote --proposal-id=1 --voter=%s --output=json %v", fooAddr, flags))
	require.Equal(t, int64(1), vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	votes := executeGetVotes(t, fmt.Sprintf("gaiacli gov query-votes --proposal-id=1 --output=json %v", flags))
	require.Len(t, votes, 1)
	require.Equal(t, int64(1), votes[0].ProposalID)
	require.Equal(t, gov.OptionYes, votes[0].Option)

	proposalsQuery = tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals --status=DepositPeriod %v", flags))
	require.Equal(t, "No matching proposals found", proposalsQuery)

	proposalsQuery = tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals --status=VotingPeriod %v", flags))
	require.Equal(t, "  1 - Test", proposalsQuery)

	// submit a second test proposal
	spStr = fmt.Sprintf("gaiacli gov submit-proposal %v", flags)
	spStr += fmt.Sprintf(" --from=%s", "foo")
	spStr += fmt.Sprintf(" --deposit=%s", "5steak")
	spStr += fmt.Sprintf(" --type=%s", "Text")
	spStr += fmt.Sprintf(" --title=%s", "Apples")
	spStr += fmt.Sprintf(" --description=%s", "test")

	executeWrite(t, spStr, pass)
	tests.WaitForNextNBlocksTM(2, port)

	proposalsQuery = tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals --latest=1 %v", flags))
	require.Equal(t, "  2 - Apples", proposalsQuery)
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

func executeGetAddrPK(t *testing.T, cmdStr string) (sdk.AccAddress, crypto.PubKey) {
	out := tests.ExecuteT(t, cmdStr)
	var ko keys.KeyOutput
	keys.UnmarshalJSON([]byte(out), &ko)

	pk, err := sdk.GetAccPubKeyBech32(ko.PubKey)
	require.NoError(t, err)

	return ko.Address, pk
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

func executeGetProposal(t *testing.T, cmdStr string) gov.Proposal {
	out := tests.ExecuteT(t, cmdStr)
	var proposal gov.Proposal
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &proposal)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return proposal
}

func executeGetVote(t *testing.T, cmdStr string) gov.Vote {
	out := tests.ExecuteT(t, cmdStr)
	var vote gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &vote)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return vote
}

func executeGetVotes(t *testing.T, cmdStr string) []gov.Vote {
	out := tests.ExecuteT(t, cmdStr)
	var votes []gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &votes)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return votes
}
