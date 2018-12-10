package clitest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/types"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/client"
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

func TestGaiaCLIMinimumFees(t *testing.T) {
	t.Parallel()
	chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server with minimum fees
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v --minimum_fees=2feeToken", gaiadHome, servAddr, p2pAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(1, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", gaiacliHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	success := executeWrite(t, fmt.Sprintf(
		"gaiacli tx send %v --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	require.False(t, success)
	cleanupDirs(gaiadHome, gaiacliHome)
}

func TestGaiaCLIFeesDeduction(t *testing.T) {
	t.Parallel()
	chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server with minimum fees
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v --minimum_fees=1fooToken", gaiadHome, servAddr, p2pAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(1, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", gaiacliHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf("fooToken").Int64())

	// test simulation
	success := executeWrite(t, fmt.Sprintf(
		"gaiacli tx send %v --amount=1000fooToken --to=%s --from=foo --fee=1fooToken --dry-run", flags, barAddr), app.DefaultKeyPass)
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, port)
	// ensure state didn't change
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf("fooToken").Int64())

	// insufficient funds (coins + fees)
	success = executeWrite(t, fmt.Sprintf(
		"gaiacli tx send %v --amount=1000fooToken --to=%s --from=foo --fee=1fooToken", flags, barAddr), app.DefaultKeyPass)
	require.False(t, success)
	tests.WaitForNextNBlocksTM(1, port)
	// ensure state didn't change
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf("fooToken").Int64())

	// test success (transfer = coins + fees)
	success = executeWrite(t, fmt.Sprintf(
		"gaiacli tx send %v --fee=300fooToken --amount=500fooToken --to=%s --from=foo", flags, barAddr), app.DefaultKeyPass)
	require.True(t, success)
	cleanupDirs(gaiadHome, gaiacliHome)
}

func TestGaiaCLISend(t *testing.T) {
	t.Parallel()
	chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", gaiadHome, servAddr, p2pAddr))
	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(1, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", gaiacliHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	executeWrite(t, fmt.Sprintf("gaiacli tx send %v --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	// Test --dry-run
	success := executeWrite(t, fmt.Sprintf("gaiacli tx send %v --amount=10%s --to=%s --from=foo --dry-run", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	require.True(t, success)
	// Check state didn't change
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	// test autosequencing
	executeWrite(t, fmt.Sprintf("gaiacli tx send %v --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(20), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(30), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	// test memo
	executeWrite(t, fmt.Sprintf("gaiacli tx send %v --amount=10%s --to=%s --from=foo --memo 'testmemo'", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(30), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(20), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	cleanupDirs(gaiadHome, gaiacliHome)
}

func TestGaiaCLIGasAuto(t *testing.T) {
	t.Parallel()
	chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", gaiadHome, servAddr, p2pAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(1, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", gaiacliHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	// Test failure with auto gas disabled and very little gas set by hand
	success := executeWrite(t, fmt.Sprintf("gaiacli tx send %v --gas=10 --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	require.False(t, success)
	tests.WaitForNextNBlocksTM(1, port)
	// Check state didn't change
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	// Test failure with negative gas
	success = executeWrite(t, fmt.Sprintf("gaiacli tx send %v --gas=-100 --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	require.False(t, success)

	// Test failure with 0 gas
	success = executeWrite(t, fmt.Sprintf("gaiacli tx send %v --gas=0 --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	require.False(t, success)

	// Enable auto gas
	success, stdout, _ := executeWriteRetStdStreams(t, fmt.Sprintf("gaiacli tx send %v --json --gas=simulate --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
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
	tests.WaitForNextNBlocksTM(1, port)
	// Check state has changed accordingly
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	cleanupDirs(gaiadHome, gaiacliHome)
}

func TestGaiaCLICreateValidator(t *testing.T) {
	t.Parallel()
	chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --chain-id=%v --node=%s", gaiacliHome, chainID, servAddr)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", gaiadHome, servAddr, p2pAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(1, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", gaiacliHome))
	consPubKey := sdk.MustBech32ifyConsPub(ed25519.GenPrivKey().PubKey())

	executeWrite(t, fmt.Sprintf("gaiacli tx send %v --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	defaultParams := stake.DefaultParams()
	initialPool := stake.InitialPool()
	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewDec(100)) // Delegate tx on GaiaAppGenState

	// create validator
	cvStr := fmt.Sprintf("gaiacli tx stake create-validator %v", flags)
	cvStr += fmt.Sprintf(" --from=%s", "bar")
	cvStr += fmt.Sprintf(" --pubkey=%s", consPubKey)
	cvStr += fmt.Sprintf(" --amount=%v", fmt.Sprintf("2%s", stakeTypes.DefaultBondDenom))
	cvStr += fmt.Sprintf(" --moniker=%v", "bar-vally")
	cvStr += fmt.Sprintf(" --commission-rate=%v", "0.05")
	cvStr += fmt.Sprintf(" --commission-max-rate=%v", "0.20")
	cvStr += fmt.Sprintf(" --commission-max-change-rate=%v", "0.10")

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
	tests.WaitForNextNBlocksTM(1, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(8), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64(), "%v", barAcc)

	validator := executeGetValidator(t, fmt.Sprintf("gaiacli query stake validator %s %v", sdk.ValAddress(barAddr), flags))
	require.Equal(t, validator.OperatorAddr, sdk.ValAddress(barAddr))
	require.True(sdk.DecEq(t, sdk.NewDec(2), validator.Tokens))

	validatorDelegations := executeGetValidatorDelegations(t, fmt.Sprintf("gaiacli query stake delegations-to %s %v", sdk.ValAddress(barAddr), flags))
	require.Len(t, validatorDelegations, 1)
	require.NotZero(t, validatorDelegations[0].Shares)

	// unbond a single share
	unbondStr := fmt.Sprintf("gaiacli tx stake unbond %v", flags)
	unbondStr += fmt.Sprintf(" --from=%s", "bar")
	unbondStr += fmt.Sprintf(" --validator=%s", sdk.ValAddress(barAddr))
	unbondStr += fmt.Sprintf(" --shares-amount=%v", "1")

	success = executeWrite(t, unbondStr, app.DefaultKeyPass)
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, port)

	/* // this won't be what we expect because we've only started unbonding, haven't completed
	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %v %v", barCech, flags))
	require.Equal(t, int64(9), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64(), "%v", barAcc)
	*/
	validator = executeGetValidator(t, fmt.Sprintf("gaiacli query stake validator %s %v", sdk.ValAddress(barAddr), flags))
	require.Equal(t, "1.0000000000", validator.Tokens.String())

	validatorUbds := executeGetValidatorUnbondingDelegations(t,
		fmt.Sprintf("gaiacli query stake unbonding-delegations-from %s %v", sdk.ValAddress(barAddr), flags))
	require.Len(t, validatorUbds, 1)
	require.Equal(t, "1", validatorUbds[0].Balance.Amount.String())

	params := executeGetParams(t, fmt.Sprintf("gaiacli query stake parameters %v", flags))
	require.True(t, defaultParams.Equal(params))

	pool := executeGetPool(t, fmt.Sprintf("gaiacli query stake pool %v", flags))
	require.Equal(t, initialPool.BondedTokens, pool.BondedTokens)
	cleanupDirs(gaiadHome, gaiacliHome)
}

func TestGaiaCLISubmitProposal(t *testing.T) {
	t.Parallel()
	chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", gaiadHome, servAddr, p2pAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(1, port)

	executeGetDepositParam(t, fmt.Sprintf("gaiacli query gov param deposit %v", flags))
	executeGetVotingParam(t, fmt.Sprintf("gaiacli query gov param voting %v", flags))
	executeGetTallyingParam(t, fmt.Sprintf("gaiacli query gov param tallying %v", flags))

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	proposalsQuery, _ := tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals %v", flags), "")
	require.Equal(t, "No matching proposals found", proposalsQuery)

	// submit a test proposal
	spStr := fmt.Sprintf("gaiacli tx gov submit-proposal %v", flags)
	spStr += fmt.Sprintf(" --from=%s", "foo")
	spStr += fmt.Sprintf(" --deposit=%s", fmt.Sprintf("5%s", stakeTypes.DefaultBondDenom))
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
	tests.WaitForNextNBlocksTM(1, port)

	txs := executeGetTxs(t, fmt.Sprintf("gaiacli query txs --tags='action:submit_proposal&proposer:%s' %v", fooAddr, flags))
	require.Len(t, txs, 1)

	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(45), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	proposal1 := executeGetProposal(t, fmt.Sprintf("gaiacli query gov proposal 1 %v", flags))
	require.Equal(t, uint64(1), proposal1.GetProposalID())
	require.Equal(t, gov.StatusDepositPeriod, proposal1.GetStatus())

	proposalsQuery, _ = tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals %v", flags), "")
	require.Equal(t, "  1 - Test", proposalsQuery)

	deposit := executeGetDeposit(t,
		fmt.Sprintf("gaiacli query gov deposit 1 %s %v", fooAddr, flags))
	require.Equal(t, int64(5), deposit.Amount.AmountOf(stakeTypes.DefaultBondDenom).Int64())

	depositStr := fmt.Sprintf("gaiacli tx gov deposit 1 %s %v", fmt.Sprintf("10%s", stakeTypes.DefaultBondDenom), flags)
	depositStr += fmt.Sprintf(" --from=%s", "foo")

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
	tests.WaitForNextNBlocksTM(1, port)

	// test query deposit
	deposits := executeGetDeposits(t, fmt.Sprintf("gaiacli query gov deposits 1 %v", flags))
	require.Len(t, deposits, 1)
	require.Equal(t, int64(15), deposits[0].Amount.AmountOf(stakeTypes.DefaultBondDenom).Int64())

	deposit = executeGetDeposit(t,
		fmt.Sprintf("gaiacli query gov deposit 1 %s %v", fooAddr, flags))
	require.Equal(t, int64(15), deposit.Amount.AmountOf(stakeTypes.DefaultBondDenom).Int64())

	txs = executeGetTxs(t, fmt.Sprintf("gaiacli query txs --tags=action:deposit&depositor:%s %v", fooAddr, flags))
	require.Len(t, txs, 1)

	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))

	require.Equal(t, int64(35), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	proposal1 = executeGetProposal(t, fmt.Sprintf("gaiacli query gov proposal 1 %v", flags))
	require.Equal(t, uint64(1), proposal1.GetProposalID())
	require.Equal(t, gov.StatusVotingPeriod, proposal1.GetStatus())

	voteStr := fmt.Sprintf("gaiacli tx gov vote 1 Yes %v", flags)
	voteStr += fmt.Sprintf(" --from=%s", "foo")

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
	tests.WaitForNextNBlocksTM(1, port)

	vote := executeGetVote(t, fmt.Sprintf("gaiacli query gov vote 1 %s %v", fooAddr, flags))
	require.Equal(t, uint64(1), vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	votes := executeGetVotes(t, fmt.Sprintf("gaiacli query gov votes 1 %v", flags))
	require.Len(t, votes, 1)
	require.Equal(t, uint64(1), votes[0].ProposalID)
	require.Equal(t, gov.OptionYes, votes[0].Option)

	txs = executeGetTxs(t, fmt.Sprintf("gaiacli query txs --tags=action:vote&voter:%s %v", fooAddr, flags))
	require.Len(t, txs, 1)

	proposalsQuery, _ = tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals --status=DepositPeriod %v", flags), "")
	require.Equal(t, "No matching proposals found", proposalsQuery)

	proposalsQuery, _ = tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals --status=VotingPeriod %v", flags), "")
	require.Equal(t, "  1 - Test", proposalsQuery)

	// submit a second test proposal
	spStr = fmt.Sprintf("gaiacli tx gov submit-proposal %v", flags)
	spStr += fmt.Sprintf(" --from=%s", "foo")
	spStr += fmt.Sprintf(" --deposit=%s", fmt.Sprintf("5%s", stakeTypes.DefaultBondDenom))
	spStr += fmt.Sprintf(" --type=%s", "Text")
	spStr += fmt.Sprintf(" --title=%s", "Apples")
	spStr += fmt.Sprintf(" --description=%s", "test")

	executeWrite(t, spStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, port)

	proposalsQuery, _ = tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals --limit=1 %v", flags), "")
	require.Equal(t, "  2 - Apples", proposalsQuery)
	cleanupDirs(gaiadHome, gaiacliHome)
}

func TestGaiaCLIValidateSignatures(t *testing.T) {
	t.Parallel()
	chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(
		t, fmt.Sprintf(
			"gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", gaiadHome, servAddr, p2pAddr,
		),
	)

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(1, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", gaiacliHome))

	// generate sendTx with default gas
	success, stdout, stderr := executeWriteRetStdStreams(
		t, fmt.Sprintf(
			"gaiacli tx send %v --amount=10%s --to=%s --from=foo --generate-only",
			flags, stakeTypes.DefaultBondDenom, barAddr,
		),
		[]string{}...,
	)
	require.True(t, success)
	require.Empty(t, stderr)

	// write  unsigned tx to file
	unsignedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// validate we can successfully sign
	success, stdout, _ = executeWriteRetStdStreams(
		t, fmt.Sprintf("gaiacli tx sign %v --name=foo %v", flags, unsignedTxFile.Name()),
		app.DefaultKeyPass,
	)
	require.True(t, success)

	stdTx := unmarshalStdTx(t, stdout)
	require.Equal(t, len(stdTx.Msgs), 1)
	require.Equal(t, 1, len(stdTx.GetSignatures()))
	require.Equal(t, fooAddr.String(), stdTx.GetSigners()[0].String())

	// write signed tx to file
	signedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// validate signatures
	success, _, _ = executeWriteRetStdStreams(
		t, fmt.Sprintf("gaiacli tx sign %v --validate-signatures %v", flags, signedTxFile.Name()),
	)
	require.True(t, success)

	// modify the transaction
	stdTx.Memo = "MODIFIED-ORIGINAL-TX-BAD"
	bz := marshalStdTx(t, stdTx)
	modSignedTxFile := writeToNewTempFile(t, string(bz))
	defer os.Remove(modSignedTxFile.Name())

	// validate signature validation failure due to different transaction sig bytes
	success, _, _ = executeWriteRetStdStreams(
		t, fmt.Sprintf("gaiacli tx sign %v --validate-signatures %v", flags, modSignedTxFile.Name()),
	)
	require.False(t, success)
}

func TestGaiaCLISendGenerateSignAndBroadcast(t *testing.T) {
	t.Parallel()
	chainID, servAddr, port, gaiadHome, gaiacliHome, p2pAddr := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", gaiacliHome, servAddr, chainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", gaiadHome, servAddr, p2pAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(1, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", gaiacliHome))

	// Test generate sendTx with default gas
	success, stdout, stderr := executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx send %v --amount=10%s --to=%s --from=foo --generate-only",
		flags, stakeTypes.DefaultBondDenom, barAddr), []string{}...)
	require.True(t, success)
	require.Empty(t, stderr)
	msg := unmarshalStdTx(t, stdout)
	require.Equal(t, msg.Fee.Gas, uint64(client.DefaultGasLimit))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test generate sendTx with --gas=$amount
	success, stdout, stderr = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx send %v --amount=10%s --to=%s --from=foo --gas=100 --generate-only",
		flags, stakeTypes.DefaultBondDenom, barAddr), []string{}...)
	require.True(t, success)
	require.Empty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.Equal(t, msg.Fee.Gas, uint64(100))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test generate sendTx, estimate gas
	success, stdout, stderr = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx send %v --amount=10%s --to=%s --from=foo --gas=simulate --generate-only",
		flags, stakeTypes.DefaultBondDenom, barAddr), []string{}...)
	require.True(t, success)
	require.NotEmpty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.True(t, msg.Fee.Gas > 0)
	require.Equal(t, len(msg.Msgs), 1)

	// Write the output to disk
	unsignedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Test sign --validate-signatures
	success, stdout, _ = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx sign %v --validate-signatures %v", flags, unsignedTxFile.Name()))
	require.False(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n 0: %v\n\nSignatures:\n\n", fooAddr.String()), stdout)

	// Test sign
	success, stdout, _ = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx sign %v --name=foo %v", flags, unsignedTxFile.Name()), app.DefaultKeyPass)
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
		"gaiacli tx sign %v --validate-signatures %v", flags, signedTxFile.Name()))
	require.True(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n 0: %v\n\nSignatures:\n 0: %v\t[OK]\n\n", fooAddr.String(),
		fooAddr.String()), stdout)

	// Test broadcast
	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	success, stdout, _ = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx broadcast %v --json %v", flags, signedTxFile.Name()))
	require.True(t, success)

	var result struct {
		Response abci.ResponseDeliverTx
	}

	require.Nil(t, app.MakeCodec().UnmarshalJSON([]byte(stdout), &result))
	require.Equal(t, msg.Fee.Gas, uint64(result.Response.GasUsed))
	require.Equal(t, msg.Fee.Gas, uint64(result.Response.GasWanted))
	tests.WaitForNextNBlocksTM(1, port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	cleanupDirs(gaiadHome, gaiacliHome)
}

func TestGaiaCLIConfig(t *testing.T) {
	t.Parallel()
	chainID, servAddr, port, gaiadHome, gaiacliHome, _ := initializeFixtures(t)
	node := fmt.Sprintf("%s:%s", servAddr, port)
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config node %s`, gaiacliHome, node))
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config output text`, gaiacliHome))
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config trust_node true`, gaiacliHome))
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config chain_id %s`, gaiacliHome, chainID))
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config trace false`, gaiacliHome))
	config, err := ioutil.ReadFile(path.Join(gaiacliHome, "config", "config.toml"))
	require.NoError(t, err)
	expectedConfig := fmt.Sprintf(`chain_id = "%s"
node = "%s"
output = "text"
trace = false
trust_node = true
`, chainID, node)
	require.Equal(t, expectedConfig, string(config))
	cleanupDirs(gaiadHome, gaiacliHome)
}

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
		"gaiad add-genesis-account %s 150%s,1000fooToken --home=%s", fooAddr, stakeTypes.DefaultBondDenom, gaiadHome))
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
