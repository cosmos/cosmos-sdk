package clitest

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestGaiaCLIKeysAddMultisig(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// key names order does not matter
	f.KeysAdd("msig1", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", keyBar, keyBaz))
	f.KeysAdd("msig2", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", keyBaz, keyBar))
	require.Equal(t, f.KeysShow("msig1").Address, f.KeysShow("msig2").Address)

	f.KeysAdd("msig3", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", keyBar, keyBaz),
		"--nosort")
	f.KeysAdd("msig4", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", keyBaz, keyBar),
		"--nosort")
	require.NotEqual(t, f.KeysShow("msig3").Address, f.KeysShow("msig4").Address)
}

func TestGaiaCLIMinimumFees(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server with minimum fees
	minGasPrice, _ := sdk.NewDecFromStr("0.000006")
	fees := fmt.Sprintf(
		"--minimum-gas-prices=%s,%s",
		sdk.NewDecCoinFromDec(feeDenom, minGasPrice),
		sdk.NewDecCoinFromDec(fee2Denom, minGasPrice),
	)
	proc := f.GDStart(fees)
	defer proc.Stop(false)

	barAddr := f.KeyAddress(keyBar)

	// Send a transaction that will get rejected
	success, _, _ := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(fee2Denom, 10))
	require.False(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure tx w/ correct fees pass
	txFees := fmt.Sprintf("--fees=%s,%s", sdk.NewInt64Coin(feeDenom, 2), sdk.NewInt64Coin(fee2Denom, 2))
	success, _, _ = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(fee2Denom, 10), txFees)
	require.True(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure tx w/ improper fees fails
	txFees = fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 5))
	success, _, _ = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 10), txFees)
	require.False(f.T, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestGaiaCLIGasPrices(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server with minimum fees
	minGasPrice, _ := sdk.NewDecFromStr("0.000006")
	proc := f.GDStart(fmt.Sprintf("--minimum-gas-prices=%s", sdk.NewDecCoinFromDec(feeDenom, minGasPrice)))
	defer proc.Stop(false)

	barAddr := f.KeyAddress(keyBar)

	// insufficient gas prices (tx fails)
	badGasPrice, _ := sdk.NewDecFromStr("0.000003")
	success, _, _ := f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 50),
		fmt.Sprintf("--gas-prices=%s", sdk.NewDecCoinFromDec(feeDenom, badGasPrice)))
	require.False(t, success)

	// wait for a block confirmation
	tests.WaitForNextNBlocksTM(1, f.Port)

	// sufficient gas prices (tx passes)
	success, _, _ = f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 50),
		fmt.Sprintf("--gas-prices=%s", sdk.NewDecCoinFromDec(feeDenom, minGasPrice)))
	require.True(t, success)

	// wait for a block confirmation
	tests.WaitForNextNBlocksTM(1, f.Port)

	f.Cleanup()
}

func TestGaiaCLIFeesDeduction(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server with minimum fees
	minGasPrice, _ := sdk.NewDecFromStr("0.000006")
	proc := f.GDStart(fmt.Sprintf("--minimum-gas-prices=%s", sdk.NewDecCoinFromDec(feeDenom, minGasPrice)))
	defer proc.Stop(false)

	// Save key addresses for later use
	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	fooAcc := f.QueryAccount(fooAddr)
	fooAmt := fooAcc.GetCoins().AmountOf(fooDenom)

	// test simulation
	success, _, _ := f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 1000),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 2)), "--dry-run")
	require.True(t, success)

	// Wait for a block
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, fooAmt.Int64(), fooAcc.GetCoins().AmountOf(fooDenom).Int64())

	// insufficient funds (coins + fees) tx fails
	success, _, _ = f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 10000000),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 2)))
	require.False(t, success)

	// Wait for a block
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, fooAmt.Int64(), fooAcc.GetCoins().AmountOf(fooDenom).Int64())

	// test success (transfer = coins + fees)
	success, _, _ = f.TxSend(
		keyFoo, barAddr, sdk.NewInt64Coin(fooDenom, 500),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(feeDenom, 2)))
	require.True(t, success)

	f.Cleanup()
}

func TestGaiaCLISend(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

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
	success, _, _ := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--dry-run")
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
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	fooAcc := f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Test failure with auto gas disabled and very little gas set by hand
	success, _, _ := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=10")
	require.False(t, success)

	// Check state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Test failure with negative gas
	success, _, _ = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=-100")
	require.False(t, success)

	// Check state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Test failure with 0 gas
	success, _, _ = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=0")
	require.False(t, success)

	// Check state didn't change
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Enable auto gas
	success, stdout, stderr := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=auto")
	require.NotEmpty(t, stderr)
	require.True(t, success)
	cdc := app.MakeCodec()
	sendResp := struct {
		Height   int64
		TxHash   string
		Response abci.ResponseDeliverTx
	}{}
	err := cdc.UnmarshalJSON([]byte(stdout), &sendResp)
	require.Nil(t, err)
	require.Equal(t, sendResp.Response.GasWanted, sendResp.Response.GasUsed)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Check state has changed accordingly
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(denom).Int64())

	f.Cleanup()
}

func TestGaiaCLICreateValidator(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

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

	defaultParams := staking.DefaultParams()
	initialPool := staking.InitialPool()
	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewInt(101)) // Delegate tx on GaiaAppGenState

	// Generate a create validator transaction and ensure correctness
	success, stdout, stderr := f.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewInt64Coin(denom, 2), "--generate-only")

	require.True(f.T, success)
	require.Empty(f.T, stderr)
	msg := unmarshalStdTx(f.T, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success, _, _ = f.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewInt64Coin(denom, 2), "--dry-run")
	require.True(t, success)

	// Create the validator
	f.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewInt64Coin(denom, 2))
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure funds were deducted properly
	barAcc = f.QueryAccount(barAddr)
	require.Equal(t, int64(8), barAcc.GetCoins().AmountOf(denom).Int64())

	// Ensure that validator state is as expected
	validator := f.QueryStakingValidator(barVal)
	require.Equal(t, validator.OperatorAddr, barVal)
	require.True(sdk.IntEq(t, sdk.NewInt(2), validator.Tokens))

	// Query delegations to the validator
	validatorDelegations := f.QueryStakingDelegationsTo(barVal)
	require.Len(t, validatorDelegations, 1)
	require.NotZero(t, validatorDelegations[0].Shares)

	// unbond a single share
	success = f.TxStakingUnbond(keyBar, "1", barVal)
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure bonded staking is correct
	validator = f.QueryStakingValidator(barVal)
	require.Equal(t, "1", validator.Tokens.String())

	// Get unbonding delegations from the validator
	validatorUbds := f.QueryStakingUnbondingDelegationsFrom(barVal)
	require.Len(t, validatorUbds, 1)
	require.Len(t, validatorUbds[0].Entries, 1)
	require.Equal(t, "1", validatorUbds[0].Entries[0].Balance.Amount.String())

	// Query staking parameters
	params := f.QueryStakingParameters()
	require.True(t, defaultParams.Equal(params))

	// Query staking pool
	pool := f.QueryStakingPool()
	require.Equal(t, initialPool.BondedTokens, pool.BondedTokens)

	f.Cleanup()
}

func TestGaiaCLISubmitProposal(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	f.QueryGovParamDeposit()
	f.QueryGovParamVoting()
	f.QueryGovParamTallying()

	fooAddr := f.KeyAddress(keyFoo)

	fooAcc := f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakingTypes.DefaultBondDenom).Int64())

	proposalsQuery := f.QueryGovProposals()
	require.Empty(t, proposalsQuery)

	// Test submit generate only for submit proposal
	success, stdout, stderr := f.TxGovSubmitProposal(
		keyFoo, "Text", "Test", "test", sdk.NewInt64Coin(denom, 5), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg := unmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success, _, _ = f.TxGovSubmitProposal(keyFoo, "Text", "Test", "test", sdk.NewInt64Coin(denom, 5), "--dry-run")
	require.True(t, success)

	// Create the proposal
	f.TxGovSubmitProposal(keyFoo, "Text", "Test", "test", sdk.NewInt64Coin(denom, 5))
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure transaction tags can be queried
	txs := f.QueryTxs(1, 50, "action:submit_proposal", fmt.Sprintf("proposer:%s", fooAddr))
	require.Len(t, txs, 1)

	// Ensure deposit was deducted
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(45), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Ensure propsal is directly queryable
	proposal1 := f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.GetProposalID())
	require.Equal(t, gov.StatusDepositPeriod, proposal1.GetStatus())

	// Ensure query proposals returns properly
	proposalsQuery = f.QueryGovProposals()
	require.Equal(t, uint64(1), proposalsQuery[0].GetProposalID())

	// Query the deposits on the proposal
	deposit := f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, int64(5), deposit.Amount.AmountOf(denom).Int64())

	// Test deposit generate only
	success, stdout, stderr = f.TxGovDeposit(1, keyFoo, sdk.NewInt64Coin(denom, 10), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Run the deposit transaction
	f.TxGovDeposit(1, keyFoo, sdk.NewInt64Coin(denom, 10))
	tests.WaitForNextNBlocksTM(1, f.Port)

	// test query deposit
	deposits := f.QueryGovDeposits(1)
	require.Len(t, deposits, 1)
	require.Equal(t, int64(15), deposits[0].Amount.AmountOf(denom).Int64())

	// Ensure querying the deposit returns the proper amount
	deposit = f.QueryGovDeposit(1, fooAddr)
	require.Equal(t, int64(15), deposit.Amount.AmountOf(denom).Int64())

	// Ensure tags are set on the transaction
	txs = f.QueryTxs(1, 50, "action:deposit", fmt.Sprintf("depositor:%s", fooAddr))
	require.Len(t, txs, 1)

	// Ensure account has expected amount of funds
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(35), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Fetch the proposal and ensure it is now in the voting period
	proposal1 = f.QueryGovProposal(1)
	require.Equal(t, uint64(1), proposal1.GetProposalID())
	require.Equal(t, gov.StatusVotingPeriod, proposal1.GetStatus())

	// Test vote generate only
	success, stdout, stderr = f.TxGovVote(1, gov.OptionYes, keyFoo, "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Vote on the proposal
	f.TxGovVote(1, gov.OptionYes, keyFoo)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Query the vote
	vote := f.QueryGovVote(1, fooAddr)
	require.Equal(t, uint64(1), vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	// Query the votes
	votes := f.QueryGovVotes(1)
	require.Len(t, votes, 1)
	require.Equal(t, uint64(1), votes[0].ProposalID)
	require.Equal(t, gov.OptionYes, votes[0].Option)

	// Ensure tags are applied to voting transaction properly
	txs = f.QueryTxs(1, 50, "action:vote", fmt.Sprintf("voter:%s", fooAddr))
	require.Len(t, txs, 1)

	// Ensure no proposals in deposit period
	proposalsQuery = f.QueryGovProposals("--status=DepositPeriod")
	require.Empty(t, proposalsQuery)

	// Ensure the proposal returns as in the voting period
	proposalsQuery = f.QueryGovProposals("--status=VotingPeriod")
	require.Equal(t, uint64(1), proposalsQuery[0].GetProposalID())

	// submit a second test proposal
	f.TxGovSubmitProposal(keyFoo, "Text", "Apples", "test", sdk.NewInt64Coin(denom, 5))
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Test limit on proposals query
	proposalsQuery = f.QueryGovProposals("--limit=1")
	require.Equal(t, uint64(2), proposalsQuery[0].GetProposalID())

	f.Cleanup()
}

func TestGaiaCLIQueryTxPagination(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	for i := 1; i <= 30; i++ {
		success := executeWrite(t, fmt.Sprintf(
			"gaiacli tx send %s --amount=%dfootoken --to=%s --from=foo",
			f.Flags(), i, barAddr), app.DefaultKeyPass)
		require.True(t, success)
		tests.WaitForNextNBlocksTM(1, f.Port)
	}

	// perPage = 15, 2 pages
	txsPage1 := f.QueryTxs(1, 15, fmt.Sprintf("sender:%s", fooAddr))
	require.Len(t, txsPage1, 15)
	txsPage2 := f.QueryTxs(2, 15, fmt.Sprintf("sender:%s", fooAddr))
	require.Len(t, txsPage2, 15)
	require.NotEqual(t, txsPage1, txsPage2)
	txsPage3 := f.QueryTxs(3, 15, fmt.Sprintf("sender:%s", fooAddr))
	require.Len(t, txsPage3, 15)
	require.Equal(t, txsPage2, txsPage3)

	// perPage = 16, 2 pages
	txsPage1 = f.QueryTxs(1, 16, fmt.Sprintf("sender:%s", fooAddr))
	require.Len(t, txsPage1, 16)
	txsPage2 = f.QueryTxs(2, 16, fmt.Sprintf("sender:%s", fooAddr))
	require.Len(t, txsPage2, 14)
	require.NotEqual(t, txsPage1, txsPage2)

	// perPage = 50
	txsPageFull := f.QueryTxs(1, 50, fmt.Sprintf("sender:%s", fooAddr))
	require.Len(t, txsPageFull, 30)
	require.Equal(t, txsPageFull, append(txsPage1, txsPage2...))

	// perPage = 0
	f.QueryTxsInvalid(errors.New("ERROR: page must greater than 0"), 0, 50, fmt.Sprintf("sender:%s", fooAddr))

	// limit = 0
	f.QueryTxsInvalid(errors.New("ERROR: limit must greater than 0"), 1, 0, fmt.Sprintf("sender:%s", fooAddr))
}

func TestGaiaCLIValidateSignatures(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	// generate sendTx with default gas
	success, stdout, stderr := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	// write  unsigned tx to file
	unsignedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// validate we can successfully sign
	success, stdout, stderr = f.TxSign(keyFoo, unsignedTxFile.Name())
	require.True(t, success)
	require.Empty(t, stderr)
	stdTx := unmarshalStdTx(t, stdout)
	require.Equal(t, len(stdTx.Msgs), 1)
	require.Equal(t, 1, len(stdTx.GetSignatures()))
	require.Equal(t, fooAddr.String(), stdTx.GetSigners()[0].String())

	// write signed tx to file
	signedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// validate signatures
	success, _, _ = f.TxSign(keyFoo, signedTxFile.Name(), "--validate-signatures")
	require.True(t, success)

	// modify the transaction
	stdTx.Memo = "MODIFIED-ORIGINAL-TX-BAD"
	bz := marshalStdTx(t, stdTx)
	modSignedTxFile := writeToNewTempFile(t, string(bz))
	defer os.Remove(modSignedTxFile.Name())

	// validate signature validation failure due to different transaction sig bytes
	success, _, _ = f.TxSign(keyFoo, modSignedTxFile.Name(), "--validate-signatures")
	require.False(t, success)

	f.Cleanup()
}

func TestGaiaCLISendGenerateSignAndBroadcast(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	// Test generate sendTx with default gas
	success, stdout, stderr := f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg := unmarshalStdTx(t, stdout)
	require.Equal(t, msg.Fee.Gas, uint64(client.DefaultGasLimit))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test generate sendTx with --gas=$amount
	success, stdout, stderr = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=100", "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.Equal(t, msg.Fee.Gas, uint64(100))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test generate sendTx, estimate gas
	success, stdout, stderr = f.TxSend(keyFoo, barAddr, sdk.NewInt64Coin(denom, 10), "--gas=auto", "--generate-only")
	require.True(t, success)
	require.NotEmpty(t, stderr)
	msg = unmarshalStdTx(t, stdout)
	require.True(t, msg.Fee.Gas > 0)
	require.Equal(t, len(msg.Msgs), 1)

	// Write the output to disk
	unsignedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Test sign --validate-signatures
	success, stdout, _ = f.TxSign(keyFoo, unsignedTxFile.Name(), "--validate-signatures")
	require.False(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n 0: %v\n\nSignatures:\n\n", fooAddr.String()), stdout)

	// Test sign
	success, stdout, _ = f.TxSign(keyFoo, unsignedTxFile.Name())
	require.True(t, success)
	msg = unmarshalStdTx(t, stdout)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 1, len(msg.GetSignatures()))
	require.Equal(t, fooAddr.String(), msg.GetSigners()[0].String())

	// Write the output to disk
	signedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Test sign --validate-signatures
	success, stdout, _ = f.TxSign(keyFoo, signedTxFile.Name(), "--validate-signatures")
	require.True(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n 0: %v\n\nSignatures:\n 0: %v\t[OK]\n\n", fooAddr.String(),
		fooAddr.String()), stdout)

	// Ensure foo has right amount of funds
	fooAcc := f.QueryAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(denom).Int64())

	// Test broadcast
	success, stdout, _ = f.TxBroadcast(signedTxFile.Name())
	require.True(t, success)

	var result struct {
		Response abci.ResponseDeliverTx
	}

	// Unmarshal the response and ensure that gas was properly used
	require.Nil(t, app.MakeCodec().UnmarshalJSON([]byte(stdout), &result))
	require.Equal(t, msg.Fee.Gas, uint64(result.Response.GasUsed))
	require.Equal(t, msg.Fee.Gas, uint64(result.Response.GasWanted))
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account state
	barAcc := f.QueryAccount(barAddr)
	fooAcc = f.QueryAccount(fooAddr)
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(denom).Int64())
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(denom).Int64())

	f.Cleanup()
}

func TestGaiaCLIMultisignInsufficientCosigners(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server with minimum fees
	proc := f.GDStart()
	defer proc.Stop(false)

	fooBarBazAddr := f.KeyAddress(keyFooBarBaz)
	barAddr := f.KeyAddress(keyBar)

	// Send some tokens from one account to the other
	success, _, _ := f.TxSend(keyFoo, fooBarBazAddr, sdk.NewInt64Coin(denom, 10))
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Test generate sendTx with multisig
	success, stdout, _ := f.TxSend(keyFooBarBaz, barAddr, sdk.NewInt64Coin(denom, 5), "--generate-only")
	require.True(t, success)

	// Write the output to disk
	unsignedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Sign with foo's key
	success, stdout, _ = f.TxSign(keyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile := writeToNewTempFile(t, stdout)
	defer os.Remove(fooSignatureFile.Name())

	// Multisign, not enough signatures
	success, stdout, _ = f.TxMultisign(unsignedTxFile.Name(), keyFooBarBaz, []string{fooSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Validate the multisignature
	success, _, _ = f.TxSign(keyFooBarBaz, signedTxFile.Name(), "--validate-signatures")
	require.False(t, success)

	// Broadcast the transaction
	success, _, _ = f.TxBroadcast(signedTxFile.Name())
	require.False(t, success)
}

func TestGaiaCLIMultisignSortSignatures(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server with minimum fees
	proc := f.GDStart()
	defer proc.Stop(false)

	fooBarBazAddr := f.KeyAddress(keyFooBarBaz)
	barAddr := f.KeyAddress(keyBar)

	// Send some tokens from one account to the other
	success, _, _ := f.TxSend(keyFoo, fooBarBazAddr, sdk.NewInt64Coin(denom, 10))
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	fooBarBazAcc := f.QueryAccount(fooBarBazAddr)
	require.Equal(t, int64(10), fooBarBazAcc.GetCoins().AmountOf(denom).Int64())

	// Test generate sendTx with multisig
	success, stdout, _ := f.TxSend(keyFooBarBaz, barAddr, sdk.NewInt64Coin(denom, 5), "--generate-only")
	require.True(t, success)

	// Write the output to disk
	unsignedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Sign with foo's key
	success, stdout, _ = f.TxSign(keyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile := writeToNewTempFile(t, stdout)
	defer os.Remove(fooSignatureFile.Name())

	// Sign with baz's key
	success, stdout, _ = f.TxSign(keyBaz, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	bazSignatureFile := writeToNewTempFile(t, stdout)
	defer os.Remove(bazSignatureFile.Name())

	// Multisign, keys in different order
	success, stdout, _ = f.TxMultisign(unsignedTxFile.Name(), keyFooBarBaz, []string{
		bazSignatureFile.Name(), fooSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Validate the multisignature
	success, _, _ = f.TxSign(keyFooBarBaz, signedTxFile.Name(), "--validate-signatures")
	require.True(t, success)

	// Broadcast the transaction
	success, _, _ = f.TxBroadcast(signedTxFile.Name())
	require.True(t, success)
}

func TestGaiaCLIMultisign(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server with minimum fees
	proc := f.GDStart()
	defer proc.Stop(false)

	fooBarBazAddr := f.KeyAddress(keyFooBarBaz)
	bazAddr := f.KeyAddress(keyBaz)

	// Send some tokens from one account to the other
	success, _, _ := f.TxSend(keyFoo, fooBarBazAddr, sdk.NewInt64Coin(denom, 10))
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	fooBarBazAcc := f.QueryAccount(fooBarBazAddr)
	require.Equal(t, int64(10), fooBarBazAcc.GetCoins().AmountOf(denom).Int64())

	// Test generate sendTx with multisig
	success, stdout, stderr := f.TxSend(keyFooBarBaz, bazAddr, sdk.NewInt64Coin(denom, 10), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	// Write the output to disk
	unsignedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Sign with foo's key
	success, stdout, _ = f.TxSign(keyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile := writeToNewTempFile(t, stdout)
	defer os.Remove(fooSignatureFile.Name())

	// Sign with bar's key
	success, stdout, _ = f.TxSign(keyBar, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	barSignatureFile := writeToNewTempFile(t, stdout)
	defer os.Remove(barSignatureFile.Name())

	// Multisign
	success, stdout, _ = f.TxMultisign(unsignedTxFile.Name(), keyFooBarBaz, []string{
		fooSignatureFile.Name(), barSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Validate the multisignature
	success, _, _ = f.TxSign(keyFooBarBaz, signedTxFile.Name(), "--validate-signatures")
	require.True(t, success)

	// Broadcast the transaction
	success, _, _ = f.TxBroadcast(signedTxFile.Name())
	require.True(t, success)
}

func TestGaiaCLIConfig(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)
	node := fmt.Sprintf("%s:%s", f.RPCAddr, f.Port)

	// Set available configuration options
	f.CLIConfig("node", node)
	f.CLIConfig("output", "text")
	f.CLIConfig("trust-node", "true")
	f.CLIConfig("chain-id", f.ChainID)
	f.CLIConfig("trace", "false")
	f.CLIConfig("indent", "true")

	config, err := ioutil.ReadFile(path.Join(f.GCLIHome, "config", "config.toml"))
	require.NoError(t, err)
	expectedConfig := fmt.Sprintf(`chain-id = "%s"
indent = true
node = "%s"
output = "text"
trace = false
trust-node = true
`, f.ChainID, node)
	require.Equal(t, expectedConfig, string(config))

	f.Cleanup()
}

func TestGaiadCollectGentxs(t *testing.T) {
	t.Parallel()
	f := NewFixtures(t)

	// Initialise temporary directories
	gentxDir, err := ioutil.TempDir("", "")
	gentxDoc := filepath.Join(gentxDir, "gentx.json")
	require.NoError(t, err)

	// Reset testing path
	f.UnsafeResetAll()

	// Initialize keys
	f.KeysDelete(keyFoo)
	f.KeysDelete(keyBar)
	f.KeysAdd(keyFoo)
	f.KeysAdd(keyBar)

	// Configure json output
	f.CLIConfig("output", "json")

	// Run init
	f.GDInit(keyFoo)

	// Add account to genesis.json
	f.AddGenesisAccount(f.KeyAddress(keyFoo), startCoins)

	// Write gentx file
	f.GenTx(keyFoo, fmt.Sprintf("--output-document=%s", gentxDoc))

	// Collect gentxs from a custom directory
	f.CollectGenTxs(fmt.Sprintf("--gentx-dir=%s", gentxDir))

	f.Cleanup(gentxDir)
}

func TestSlashingGetParams(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	params := f.QuerySlashingParams()
	require.Equal(t, time.Duration(120000000000), params.MaxEvidenceAge)
	require.Equal(t, int64(100), params.SignedBlocksWindow)
	require.Equal(t, sdk.NewDecWithPrec(5, 1), params.MinSignedPerWindow)
}
