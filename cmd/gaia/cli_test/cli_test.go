package clitest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/slashing"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
)

func TestGaiaCLIMinimumFees(t *testing.T) {
	t.Parallel()
	f := initializeFixtures(t)

	// start gaiad server with minimum fees
	fees := fmt.Sprintf("--minimum_fees=%s,%s", sdk.NewInt64Coin(feeDenom, 2), sdk.NewInt64Coin(denom, 2))
	proc := f.GDStart(fees)
	defer proc.Stop(false)

	barAddr := f.KeyAddress(keyBar)
	// fooAddr := f.KeyAddress(keyFoo)

	// Check the amount of coins in the foo account to ensure that the right amount exists
	fooAcc := f.QueryAccount(f.KeyAddress(keyFoo))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Send a transaction that will get rejected
	success := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10))
	require.False(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// TODO: Make this work
	// // Ensure tx w/ correct fees (stake) pass
	// txFees := fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(denom, 23))
	// success = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), txFees)
	// require.True(f.T, success)
	// tests.WaitForNextNBlocksTM(1, f.Port)
	//
	// // Ensure tx w/ correct fees (feetoken) pass
	// txFees = fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 23))
	// success = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(feeDenom, 10), txFees)
	// require.True(f.T, success)
	// tests.WaitForNextNBlocksTM(1, f.Port)
	//
	// // Ensure tx w/ improper fees (footoken) fails
	// txFees = fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(fooDenom, 23))
	// success = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 10), txFees)
	// require.False(f.T, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestGaiaCLIFeesDeduction(t *testing.T) {
	t.Parallel()
	f := initializeFixtures(t)

	// start gaiad server with minimum fees
	proc := f.GDStart(fmt.Sprintf("--minimum_fees=%s", sdk.NewInt64Coin(fooDenom, 1)))
	defer proc.Stop(false)

	// Save key addresses for later use
	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	fooAcc := f.QueryAccount(fooAddr)
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf(fooDenom).Int64())

	// test simulation
	success := f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 1000),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(fooDenom, 1)), "--dry-run")
	require.True(t, success)

	// Wait for a block
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf(fooDenom).Int64())

	// insufficient funds (coins + fees) tx fails
	success = f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 1000),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(fooDenom, 1)))
	require.False(t, success)

	// Wait for a block
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf(fooDenom).Int64())

	// test success (transfer = coins + fees)
	success = f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 500),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(fooDenom, 300)))
	require.True(t, success)

	f.Cleanup()
}

func TestGaiaCLISend(t *testing.T) {
	t.Parallel()
	f := initializeFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	// Save key addresses for later use
	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	fooAcc := f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Send some tokens from one account to the other
	f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10))
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	barAcc := f.QueryAccount(barAddr)
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(denom).Int64())
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Test --dry-run
	success := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--dry-run")
	require.True(t, success)

	// Check state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(denom).Int64())

	// test autosequencing
	f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10))
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	barAcc = f.QueryAccount(barAddr)
	require.Equal(t, int64(20), barAcc.GetCoins().AmountOf(denom).Int64())
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(30), fooAcc.GetCoins().AmountOf(denom).Int64())

	// test memo
	f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--memo='testmemo'")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	barAcc = f.QueryAccount(barAddr)
	require.Equal(t, int64(30), barAcc.GetCoins().AmountOf(denom).Int64())
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(20), fooAcc.GetCoins().AmountOf(denom).Int64())

	f.Cleanup()
}

func TestGaiaCLIGasAuto(t *testing.T) {
	t.Parallel()
	f := initializeFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	fooAcc := f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Test failure with auto gas disabled and very little gas set by hand
	success := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=10")
	require.False(t, success)

	// Check state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Test failure with negative gas
	success = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=-100")
	require.False(t, success)

	// Check state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Test failure with 0 gas
	success = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=0")
	require.False(t, success)

	// Check state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Enable auto gas
	sendResp := f.TxSendWResponse(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=auto")
	require.Equal(t, sendResp.Response.GasWanted, sendResp.Response.GasUsed)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Check state has changed accordingly
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(denom).Int64())

	f.Cleanup()
}

func TestGaiaCLICreateValidator(t *testing.T) {
	t.Parallel()
	f := initializeFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)
	barVal := sdk.ValAddress(barAddr)

	consPubKey := sdk.MustBech32ifyConsPub(ed25519.GenPrivKey().PubKey())

	f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10))
	tests.WaitForNextNBlocksTM(1, f.Port)

	barAcc := f.QueryAccount(barAddr)
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(denom).Int64())
	fooAcc := f.QueryAccount(fooAddr)
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(denom).Int64())

	defaultParams := stake.DefaultParams()
	initialPool := stake.InitialPool()
	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewInt(101)) // Delegate tx on GaiaAppGenState

	// Generate a create validator transaction and ensure correctness
	success, stdout, stderr := f.TxStakeCreateValidator(keyBar, consPubKey, sdk.NewInt64Coin(denom, 2), "--generate-only")

	require.True(f.T, success)
	require.Empty(f.T, stderr)
	msg := unmarshalStdTx(f.T, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success, _, _ = f.TxStakeCreateValidator(keyBar, consPubKey, sdk.NewInt64Coin(denom, 2), "--dry-run")
	require.True(t, success)

	// Create the validator
	f.TxStakeCreateValidator(keyBar, consPubKey, sdk.NewInt64Coin(denom, 2))
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure funds were deducted properly
	barAcc = f.QueryAccount(barAddr)
	require.Equal(t, int64(8), barAcc.GetCoins().AmountOf(denom).Int64())

	// Ensure that validator state is as expected
	validator := f.QueryStakeValidator(barVal)
	require.Equal(t, validator.OperatorAddr, barVal)
	require.True(sdk.IntEq(t, sdk.NewInt(2), validator.Tokens))

	// Query delegations to the validator
	validatorDelegations := f.QueryStakeDelegationsTo(barVal)
	require.Len(t, validatorDelegations, 1)
	require.NotZero(t, validatorDelegations[0].Shares)

	// unbond a single share
	success = f.TxStakeUnbond(keyBar, "1", barVal)
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure bonded stake is correct
	validator = f.QueryStakeValidator(barVal)
	require.Equal(t, "1", validator.Tokens.String())

	// Get unbonding delegations from the validator
	validatorUbds := f.QueryStakeUnbondingDelegationsFrom(barVal)
	require.Len(t, validatorUbds, 1)
	require.Equal(t, "1", validatorUbds[0].Balance.Amount.String())

	// Query staking parameters
	params := f.QueryStakeParameters()
	require.True(t, defaultParams.Equal(params))

	// Query staking pool
	pool := f.QueryStakePool()
	require.Equal(t, initialPool.BondedTokens, pool.BondedTokens)

	f.Cleanup()
}

func TestGaiaCLISubmitProposal(t *testing.T) {
	t.Parallel()
	f := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", f.GCLIHome, f.RPCAddr, f.ChainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", f.GDHome, f.RPCAddr, f.P2PAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(f.Port)
	tests.WaitForNextNBlocksTM(1, f.Port)

	executeGetDepositParam(t, fmt.Sprintf("gaiacli query gov param deposit %v", flags))
	executeGetVotingParam(t, fmt.Sprintf("gaiacli query gov param voting %v", flags))
	executeGetTallyingParam(t, fmt.Sprintf("gaiacli query gov param tallying %v", flags))

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", f.GCLIHome))

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
	tests.WaitForNextNBlocksTM(1, f.Port)

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
	tests.WaitForNextNBlocksTM(1, f.Port)

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
	tests.WaitForNextNBlocksTM(1, f.Port)

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
	tests.WaitForNextNBlocksTM(1, f.Port)

	proposalsQuery, _ = tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals --limit=1 %v", flags), "")
	require.Equal(t, "  2 - Apples", proposalsQuery)
	cleanupDirs(f.GDHome, f.GCLIHome)
}

func TestGaiaCLIValidateSignatures(t *testing.T) {
	t.Parallel()
	f := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", f.GCLIHome, f.RPCAddr, f.ChainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(
		t, fmt.Sprintf(
			"gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", f.GDHome, f.RPCAddr, f.P2PAddr,
		),
	)

	defer proc.Stop(false)
	tests.WaitForTMStart(f.Port)
	tests.WaitForNextNBlocksTM(1, f.Port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", f.GCLIHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", f.GCLIHome))

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
	f := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", f.GCLIHome, f.RPCAddr, f.ChainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v", f.GDHome, f.RPCAddr, f.P2PAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(f.Port)
	tests.WaitForNextNBlocksTM(1, f.Port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", f.GCLIHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", f.GCLIHome))

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
		"gaiacli tx send %v --amount=10%s --to=%s --from=foo --gas=auto --generate-only",
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
	tests.WaitForNextNBlocksTM(1, f.Port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	cleanupDirs(f.GDHome, f.GCLIHome)
}

func TestGaiaCLIConfig(t *testing.T) {
	t.Parallel()
	f := initializeFixtures(t)
	node := fmt.Sprintf("%s:%s", f.RPCAddr, f.Port)
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config node %s`, f.GCLIHome, node))
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config output text`, f.GCLIHome))
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config trust-node true`, f.GCLIHome))
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config chain-id %s`, f.GCLIHome, f.ChainID))
	executeWrite(t, fmt.Sprintf(`gaiacli --home=%s config trace false`, f.GCLIHome))
	config, err := ioutil.ReadFile(path.Join(f.GCLIHome, "config", "config.toml"))
	require.NoError(t, err)
	expectedConfig := fmt.Sprintf(`chain-id = "%s"
node = "%s"
output = "text"
trace = false
trust-node = true
`, f.ChainID, node)
	require.Equal(t, expectedConfig, string(config))
	cleanupDirs(f.GDHome, f.GCLIHome)
}

func TestGaiadCollectGentxs(t *testing.T) {
	t.Parallel()
	// Initialise temporary directories
	gaiadHome, gaiacliHome := getTestingHomeDirs(t.Name())
	gentxDir, err := ioutil.TempDir("", "")
	gentxDoc := filepath.Join(gentxDir, "gentx.json")
	require.NoError(t, err)

	tests.ExecuteT(t, fmt.Sprintf("gaiad --home=%s unsafe-reset-all", gaiadHome), "")
	os.RemoveAll(filepath.Join(gaiadHome, "config", "gentx"))
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s foo", gaiacliHome), app.DefaultKeyPass)
	executeWrite(t, fmt.Sprintf("gaiacli keys delete --home=%s bar", gaiacliHome), app.DefaultKeyPass)
	executeWriteCheckErr(t, fmt.Sprintf("gaiacli keys add --home=%s foo", gaiacliHome), app.DefaultKeyPass)
	executeWriteCheckErr(t, fmt.Sprintf("gaiacli keys add --home=%s bar", gaiacliHome), app.DefaultKeyPass)
	executeWriteCheckErr(t, fmt.Sprintf("gaiacli config --home=%s output json", gaiacliHome))
	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", gaiacliHome))

	// Run init
	_ = executeInit(t, fmt.Sprintf("gaiad init -o --moniker=foo --home=%s", gaiadHome))
	// Add account to genesis.json
	executeWriteCheckErr(t, fmt.Sprintf(
		"gaiad add-genesis-account %s 150%s,1000footoken --home=%s", fooAddr, stakeTypes.DefaultBondDenom, gaiadHome))
	executeWrite(t, fmt.Sprintf("cat %s%sconfig%sgenesis.json", gaiadHome, string(os.PathSeparator), string(os.PathSeparator)))
	// Write gentx file
	executeWriteCheckErr(t, fmt.Sprintf(
		"gaiad gentx --name=foo --home=%s --home-client=%s --output-document=%s", gaiadHome, gaiacliHome, gentxDoc), app.DefaultKeyPass)
	// Collect gentxs from a custom directory
	executeWriteCheckErr(t, fmt.Sprintf("gaiad collect-gentxs --home=%s --gentx-dir=%s", gaiadHome, gentxDir), app.DefaultKeyPass)
	cleanupDirs(gaiadHome, gaiacliHome, gentxDir)
}

// ---------------------------------------------------------------------------
// Slashing

func TestSlashingGetParams(t *testing.T) {
	t.Parallel()
	cdc := app.MakeCodec()
	f := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", f.GCLIHome, f.RPCAddr, f.ChainID)

	// start gaiad server
	proc := tests.GoExecuteTWithStdout(
		t,
		fmt.Sprintf(
			"gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v",
			f.GDHome, f.RPCAddr, f.P2PAddr,
		),
	)

	defer proc.Stop(false)
	tests.WaitForTMStart(f.Port)
	tests.WaitForNextNBlocksTM(1, f.Port)

	res, errStr := tests.ExecuteT(t, fmt.Sprintf("gaiacli query slashing params %s", flags), "")
	require.Empty(t, errStr)

	var params slashing.Params
	err := cdc.UnmarshalJSON([]byte(res), &params)
	require.NoError(t, err)
}
