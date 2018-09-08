// +build cli_test

package clitest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/client"
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
	gaiadHome   = ""
	gaiacliHome = ""
)

func init() {
	gaiadHome, gaiacliHome = getTestingHomeDirs()
}

func TestGaiaCLISend(t *testing.T) {
	chainID, servAddr, port := initializeFixtures(t)
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

	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%s --from=foo", flags, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf("steak").Int64())

	// Test --dry-run
	success := executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%s --from=foo --dry-run", flags, barAddr), app.DefaultKeyPass)
	require.True(t, success)
	// Check state didn't change
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf("steak").Int64())

	// test autosequencing
	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%s --from=foo", flags, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(20), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(30), fooAcc.GetCoins().AmountOf("steak").Int64())

	// test memo
	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%s --from=foo --memo 'testmemo'", flags, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(30), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(20), fooAcc.GetCoins().AmountOf("steak").Int64())
}

func TestGaiaCLIGasAuto(t *testing.T) {
	chainID, servAddr, port := initializeFixtures(t)
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

	// Test failure with auto gas disabled and very little gas set by hand
	success := executeWrite(t, fmt.Sprintf("gaiacli send %v --gas=10 --amount=10steak --to=%s --from=foo", flags, barAddr), app.DefaultKeyPass)
	require.False(t, success)
	tests.WaitForNextNBlocksTM(2, port)
	// Check state didn't change
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf("steak").Int64())

	// Enable auto gas
	success, stdout, _ := executeWriteRetStdStreams(t, fmt.Sprintf("gaiacli send %v --json --gas=0 --amount=10steak --to=%s --from=foo", flags, barAddr), app.DefaultKeyPass)
	require.True(t, success)
	// check that gas wanted == gas used
	cdc := app.MakeCodec()
	jsonOutput := struct {
		Height   int64
		TxHash   string
		Response abci.ResponseDeliverTx
	}{}
	require.Nil(t, cdc.UnmarshalJSON([]byte(stdout), &jsonOutput))
	require.Equal(t, jsonOutput.Response.GasWanted, jsonOutput.Response.GasUsed)
	tests.WaitForNextNBlocksTM(2, port)
	// Check state has changed accordingly
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf("steak").Int64())
}

func TestGaiaCLICreateValidator(t *testing.T) {
	chainID, servAddr, port := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v", gaiadHome, servAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(2, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --output=json --home=%s", gaiacliHome))
	barAddr, barPubKey := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --output=json --home=%s", gaiacliHome))
	barCeshPubKey := sdk.MustBech32ifyConsPub(barPubKey)

	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10steak --to=%s --from=foo", flags, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf("steak").Int64())
	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf("steak").Int64())

	defaultParams := stake.DefaultParams()
	initialPool := stake.InitialPool()
	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewDec(100)) // Delegate tx on GaiaAppGenState
	initialPool = initialPool.ProcessProvisions(defaultParams)               // provisions are added to the pool every hour

	// create validator
	cvStr := fmt.Sprintf("gaiacli stake create-validator %v", flags)
	cvStr += fmt.Sprintf(" --from=%s", "bar")
	cvStr += fmt.Sprintf(" --pubkey=%s", barCeshPubKey)
	cvStr += fmt.Sprintf(" --amount=%v", "2steak")
	cvStr += fmt.Sprintf(" --moniker=%v", "bar-vally")

	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewDec(1))

	// Test --generate-only
	success, stdout, stderr := executeWriteRetStdStreams(t, cvStr+" --generate-only", app.DefaultKeyPass)
	require.True(t, success)
	require.True(t, success)
	require.Empty(t, stderr)
	msg := unmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success = executeWrite(t, cvStr+" --dry-run", app.DefaultKeyPass)
	require.True(t, success)

	executeWrite(t, cvStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", barAddr, flags))
	require.Equal(t, int64(8), barAcc.GetCoins().AmountOf("steak").Int64(), "%v", barAcc)

	validator := executeGetValidator(t, fmt.Sprintf("gaiacli stake validator %s --output=json %v", sdk.ValAddress(barAddr), flags))
	require.Equal(t, validator.Operator, sdk.ValAddress(barAddr))
	require.True(sdk.DecEq(t, sdk.NewDec(2), validator.Tokens))

	// unbond a single share
	unbondStr := fmt.Sprintf("gaiacli stake unbond begin %v", flags)
	unbondStr += fmt.Sprintf(" --from=%s", "bar")
	unbondStr += fmt.Sprintf(" --validator=%s", sdk.ValAddress(barAddr))
	unbondStr += fmt.Sprintf(" --shares-amount=%v", "1")

	success = executeWrite(t, unbondStr, app.DefaultKeyPass)
	require.True(t, success)
	tests.WaitForNextNBlocksTM(2, port)

	/* // this won't be what we expect because we've only started unbonding, haven't completed
	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barCech, flags))
	require.Equal(t, int64(9), barAcc.GetCoins().AmountOf("steak").Int64(), "%v", barAcc)
	*/
	validator = executeGetValidator(t, fmt.Sprintf("gaiacli stake validator %s --output=json %v", sdk.ValAddress(barAddr), flags))
	require.Equal(t, "1.0000000000", validator.Tokens.String())

	params := executeGetParams(t, fmt.Sprintf("gaiacli stake parameters --output=json %v", flags))
	require.True(t, defaultParams.Equal(params))

	pool := executeGetPool(t, fmt.Sprintf("gaiacli stake pool --output=json %v", flags))
	require.Equal(t, initialPool.DateLastCommissionReset, pool.DateLastCommissionReset)
	require.Equal(t, initialPool.PrevBondedShares, pool.PrevBondedShares)
	require.Equal(t, initialPool.BondedTokens, pool.BondedTokens)
}

func TestGaiaCLISubmitProposal(t *testing.T) {
	chainID, servAddr, port := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v", gaiadHome, servAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(2, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --output=json --home=%s", gaiacliHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf("steak").Int64())

	proposalsQuery := tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals %v", flags), "")
	require.Equal(t, "No matching proposals found", proposalsQuery)

	// submit a test proposal
	spStr := fmt.Sprintf("gaiacli gov submit-proposal %v", flags)
	spStr += fmt.Sprintf(" --from=%s", "foo")
	spStr += fmt.Sprintf(" --deposit=%s", "5steak")
	spStr += fmt.Sprintf(" --type=%s", "Text")
	spStr += fmt.Sprintf(" --title=%s", "Test")
	spStr += fmt.Sprintf(" --description=%s", "test")

	// Test generate only
	success, stdout, stderr := executeWriteRetStdStreams(t, spStr+" --generate-only", app.DefaultKeyPass)
	require.True(t, success)
	require.True(t, success)
	require.Empty(t, stderr)
	msg := unmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success = executeWrite(t, spStr+" --dry-run", app.DefaultKeyPass)
	require.True(t, success)

	executeWrite(t, spStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %s %v", fooAddr, flags))
	require.Equal(t, int64(45), fooAcc.GetCoins().AmountOf("steak").Int64())

	proposal1 := executeGetProposal(t, fmt.Sprintf("gaiacli gov query-proposal --proposal-id=1 --output=json %v", flags))
	require.Equal(t, int64(1), proposal1.GetProposalID())
	require.Equal(t, gov.StatusDepositPeriod, proposal1.GetStatus())

	proposalsQuery = tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals %v", flags), "")
	require.Equal(t, "  1 - Test", proposalsQuery)

	depositStr := fmt.Sprintf("gaiacli gov deposit %v", flags)
	depositStr += fmt.Sprintf(" --from=%s", "foo")
	depositStr += fmt.Sprintf(" --deposit=%s", "10steak")
	depositStr += fmt.Sprintf(" --proposal-id=%s", "1")

	// Test generate only
	success, stdout, stderr = executeWriteRetStdStreams(t, depositStr+" --generate-only", app.DefaultKeyPass)
	require.True(t, success)
	require.True(t, success)
	require.Empty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	executeWrite(t, depositStr, app.DefaultKeyPass)
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

	// Test generate only
	success, stdout, stderr = executeWriteRetStdStreams(t, voteStr+" --generate-only", app.DefaultKeyPass)
	require.True(t, success)
	require.True(t, success)
	require.Empty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	executeWrite(t, voteStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	vote := executeGetVote(t, fmt.Sprintf("gaiacli gov query-vote --proposal-id=1 --voter=%s --output=json %v", fooAddr, flags))
	require.Equal(t, int64(1), vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	votes := executeGetVotes(t, fmt.Sprintf("gaiacli gov query-votes --proposal-id=1 --output=json %v", flags))
	require.Len(t, votes, 1)
	require.Equal(t, int64(1), votes[0].ProposalID)
	require.Equal(t, gov.OptionYes, votes[0].Option)

	proposalsQuery = tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals --status=DepositPeriod %v", flags), "")
	require.Equal(t, "No matching proposals found", proposalsQuery)

	proposalsQuery = tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals --status=VotingPeriod %v", flags), "")
	require.Equal(t, "  1 - Test", proposalsQuery)

	// submit a second test proposal
	spStr = fmt.Sprintf("gaiacli gov submit-proposal %v", flags)
	spStr += fmt.Sprintf(" --from=%s", "foo")
	spStr += fmt.Sprintf(" --deposit=%s", "5steak")
	spStr += fmt.Sprintf(" --type=%s", "Text")
	spStr += fmt.Sprintf(" --title=%s", "Apples")
	spStr += fmt.Sprintf(" --description=%s", "test")

	executeWrite(t, spStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	proposalsQuery = tests.ExecuteT(t, fmt.Sprintf("gaiacli gov query-proposals --latest=1 %v", flags), "")
	require.Equal(t, "  2 - Apples", proposalsQuery)
}

func TestGaiaCLISendGenerateAndSign(t *testing.T) {
	chainID, servAddr, port := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v", gaiadHome, servAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(2, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --output=json --home=%s", gaiacliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --output=json --home=%s", gaiacliHome))

	// Test generate sendTx with default gas
	success, stdout, stderr := executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli send %v --amount=10steak --to=%s --from=foo --generate-only",
		flags, barAddr), []string{}...)
	require.True(t, success)
	require.Empty(t, stderr)
	msg := unmarshalStdTx(t, stdout)
	require.Equal(t, msg.Fee.Gas, int64(client.DefaultGasLimit))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test generate sendTx, estimate gas
	success, stdout, stderr = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli send %v --amount=10steak --to=%s --from=foo --gas=0 --generate-only",
		flags, barAddr), []string{}...)
	require.True(t, success)
	require.NotEmpty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)

	// Test generate sendTx with --gas=$amount
	success, stdout, stderr = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli send %v --amount=10steak --to=%s --from=foo --gas=100 --generate-only",
		flags, barAddr), []string{}...)
	require.True(t, success)
	require.Empty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.Equal(t, msg.Fee.Gas, int64(100))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Write the output to disk
	unsignedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Test sign --print-sigs
	success, stdout, _ = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli sign %v --print-sigs %v", flags, unsignedTxFile.Name()))
	require.True(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n 0: %v\n\nSignatures:\n", fooAddr.String()), stdout)

	// Test sign
	success, stdout, _ = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli sign %v --name=foo %v", flags, unsignedTxFile.Name()), app.DefaultKeyPass)
	require.True(t, success)
	msg = unmarshalStdTx(t, stdout)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 1, len(msg.GetSignatures()))
	require.Equal(t, fooAddr.String(), msg.GetSigners()[0].String())

	// Write the output to disk
	signedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Test sign --print-signatures
	success, stdout, _ = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli sign %v --print-sigs %v", flags, signedTxFile.Name()))
	require.True(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n 0: %v\n\nSignatures:\n 0: %v\n", fooAddr.String(), fooAddr.String()), stdout)
}

//___________________________________________________________________________________
// helper methods

func getTestingHomeDirs() (string, string) {
	tmpDir := os.TempDir()
	gaiadHome := fmt.Sprintf("%s%s.test_gaiad", tmpDir, string(os.PathSeparator))
	gaiacliHome := fmt.Sprintf("%s%s.test_gaiacli", tmpDir, string(os.PathSeparator))
	return gaiadHome, gaiacliHome
}

func initializeFixtures(t *testing.T) (chainID, servAddr, port string) {
	tests.ExecuteT(t, fmt.Sprintf("gaiad --home=%s unsafe-reset-all", gaiadHome), "")
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s foo", gaiacliHome), app.DefaultKeyPass)
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s bar", gaiacliHome), app.DefaultKeyPass)

	chainID = executeInit(t, fmt.Sprintf("gaiad init -o --name=foo --home=%s --home-client=%s", gaiadHome, gaiacliHome))
	executeWrite(t, fmt.Sprintf("gaiacli keys add --home=%s bar", gaiacliHome), app.DefaultKeyPass)

	// get a free port, also setup some common flags
	servAddr, port, err := server.FreeTCPAddr()
	require.NoError(t, err)
	return
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

//___________________________________________________________________________________
// executors

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
	out := tests.ExecuteT(t, cmdStr, app.DefaultKeyPass)

	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(t, err)

	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(t, err)

	return
}

func executeGetAddrPK(t *testing.T, cmdStr string) (sdk.AccAddress, crypto.PubKey) {
	out := tests.ExecuteT(t, cmdStr, "")
	var ko keys.KeyOutput
	keys.UnmarshalJSON([]byte(out), &ko)

	pk, err := sdk.GetAccPubKeyBech32(ko.PubKey)
	require.NoError(t, err)

	accAddr, err := sdk.AccAddressFromBech32(ko.Address)
	require.NoError(t, err)

	return accAddr, pk
}

func executeGetAccount(t *testing.T, cmdStr string) auth.BaseAccount {
	out := tests.ExecuteT(t, cmdStr, "")
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

//___________________________________________________________________________________
// stake

func executeGetValidator(t *testing.T, cmdStr string) stake.Validator {
	out := tests.ExecuteT(t, cmdStr, "")
	var validator stake.Validator
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &validator)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return validator
}

func executeGetPool(t *testing.T, cmdStr string) stake.Pool {
	out := tests.ExecuteT(t, cmdStr, "")
	var pool stake.Pool
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &pool)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return pool
}

func executeGetParams(t *testing.T, cmdStr string) stake.Params {
	out := tests.ExecuteT(t, cmdStr, "")
	var params stake.Params
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &params)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return params
}

//___________________________________________________________________________________
// gov

func executeGetProposal(t *testing.T, cmdStr string) gov.Proposal {
	out := tests.ExecuteT(t, cmdStr, "")
	var proposal gov.Proposal
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &proposal)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return proposal
}

func executeGetVote(t *testing.T, cmdStr string) gov.Vote {
	out := tests.ExecuteT(t, cmdStr, "")
	var vote gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &vote)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return vote
}

func executeGetVotes(t *testing.T, cmdStr string) []gov.Vote {
	out := tests.ExecuteT(t, cmdStr, "")
	var votes []gov.Vote
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &votes)
	require.NoError(t, err, "out %v\n, err %v", out, err)
	return votes
}
