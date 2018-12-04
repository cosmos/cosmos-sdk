package clitest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestGaiaCLIMinimumFees(t *testing.T) {
	// t.Parallel()
	f := InitFixtures(t)

	// start gaiad server with minimum fees and wait for start
	proc := f.StartGaiad(t, "--minimum_fees=2feeToken")
	defer proc.Stop(false)

	// Pull the two addresses
	fooAddr := f.GetKeyAddress(keyFoo)
	barAddr := f.GetKeyAddress(keyBar)

	// Get account balance and ensure correctness
	fooAcc := f.GetAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Send should fail because of min_fees
	success := f.sendCoins(t, bondDenomTokens(10), keyFoo, barAddr)
	require.False(t, success)

	// Cleanup on success
	f.CleanupDirs()
}

func TestGaiaCLIFeesDeduction(t *testing.T) {
	// t.Parallel()
	f := InitFixtures(t)

	// Start gaiad with proper args and wait for start
	proc := f.StartGaiad(t)
	t.Log("Starting gaiad")
	defer proc.Stop(false)

	// Pull the two addresses
	fooAddr := f.GetKeyAddress(keyFoo)
	barAddr := f.GetKeyAddress(keyBar)

	// Account balance should exist
	fooAcc := f.GetAccount(fooAddr)
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf(tokenFoo).Int64())

	// test simulation
	success := f.sendCoins(t, fooTokens(1000),
		keyFoo, barAddr, feeFlag(fooTokens(1)), "--dry-run")
	require.True(t, success)

	// Wait a block for tx to be included
	tests.WaitForNextNBlocksTM(1, f.port)

	// ensure state didn't change
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf(tokenFoo).Int64())

	// A send w/o enough funds in the wallet should fail
	success = f.sendCoins(t, fooTokens(1000),
		keyFoo, barAddr, feeFlag(fooTokens(1)))
	require.False(t, success)

	// Wait a block for tx to be included
	tests.WaitForNextNBlocksTM(1, f.port)

	// ensure state didn't change
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf(tokenFoo).Int64())

	// test success (transfer = coins + fees)
	success = f.sendCoins(t, fooTokens(500), keyFoo, barAddr, feeFlag(fooTokens(300)))
	require.True(t, success)

	// Cleanup on success
	f.CleanupDirs()
}

func TestGaiaCLISend(t *testing.T) {
	// t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.StartGaiad(t)
	t.Log("Starting gaiad")
	defer proc.Stop(false)

	// Pull the two addresses
	fooAddr := f.GetKeyAddress(keyFoo)
	barAddr := f.GetKeyAddress(keyBar)

	// Ensure foo has enough coins
	fooAcc := f.GetAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Send some coins to bar
	f.sendCoins(t, bondDenomTokens(10), keyFoo, barAddr)
	tests.WaitForNextNBlocksTM(1, f.port)

	// Check balances of both foo and bar
	barAcc := f.GetAccount(barAddr)
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(bondDenom).Int64())
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Test --dry-run
	success := f.sendCoins(t, bondDenomTokens(10), keyFoo, barAddr, "--dry-run")
	require.True(t, success)

	// Check state didn't change
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// test autosequencing
	f.sendCoins(t, bondDenomTokens(10), keyFoo, barAddr)
	tests.WaitForNextNBlocksTM(1, f.port)

	// Check account balances correct
	barAcc = f.GetAccount(barAddr)
	require.Equal(t, int64(20), barAcc.GetCoins().AmountOf(bondDenom).Int64())
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(30), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// test memo
	f.sendCoins(t, bondDenomTokens(10), keyFoo, barAddr, "--memo='testmemo'")
	tests.WaitForNextNBlocksTM(1, f.port)

	// Ensure balances are correct
	barAcc = f.GetAccount(barAddr)
	require.Equal(t, int64(30), barAcc.GetCoins().AmountOf(bondDenom).Int64())
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(20), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Cleanup on success
	f.CleanupDirs()
}

func TestGaiaCLIGasAuto(t *testing.T) {
	// t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.StartGaiad(t)
	t.Log("Starting gaiad")

	defer proc.Stop(false)
	cdc := app.MakeCodec()

	// Pull the two addresses
	fooAddr := f.GetKeyAddress(keyFoo)
	barAddr := f.GetKeyAddress(keyBar)

	// Ensure balance is correct
	fooAcc := f.GetAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Test failure with auto gas disabled and very little gas set by hand
	success := f.sendCoins(t, bondDenomTokens(10), keyFoo, barAddr, "--gas=10")
	require.False(t, success)
	tests.WaitForNextNBlocksTM(1, f.port)

	// Check state didn't change
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Test failure with negative gas
	success = f.sendCoins(t, bondDenomTokens(10), keyFoo, barAddr, "--gas=-100")
	require.False(t, success)

	// Test failure with 0 gas
	success = f.sendCoins(t, bondDenomTokens(10), keyFoo, barAddr, "--gas=0")
	require.False(t, success)

	// Simulate sending coins
	response := f.sendCoinsResponse(t, bondDenomTokens(10), keyFoo, barAddr, "--gas=simulate")
	res, err := cdc.MarshalJSON(response)
	require.NoError(t, err)

	// Unmarshal response and ensure GasWanted == GasUsed
	result := DeliverTxResponse{}
	require.Nil(t, cdc.UnmarshalJSON([]byte(res), &result))
	require.Equal(t, result.Response.GasWanted, result.Response.GasUsed)
	tests.WaitForNextNBlocksTM(1, f.port)

	// Check state has changed accordingly
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Cleanup on success
	f.CleanupDirs()
}

func TestGaiaCLICreateValidator(t *testing.T) {
	// t.Parallel()
	f := InitFixtures(t)

	// Start gaiad server
	proc := f.StartGaiad(t)
	t.Log("Starting gaiad")
	defer proc.Stop(false)

	// Pull the two addresses
	fooAddr := f.GetKeyAddress(keyFoo)
	barAddr := f.GetKeyAddress(keyBar)

	// Generate a new consensus public key
	consPubKey := sdk.MustBech32ifyConsPub(ed25519.GenPrivKey().PubKey())

	// Send some coins to bar
	f.sendCoins(t, bondDenomTokens(10), keyFoo, barAddr)
	tests.WaitForNextNBlocksTM(1, f.port)

	// Ensure account balances
	barAcc := f.GetAccount(barAddr)
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(bondDenom).Int64())
	fooAcc := f.GetAccount(fooAddr)
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Initalize parameters
	defaultParams := stake.DefaultParams()
	initialPool := stake.InitialPool()
	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewDec(100)) // Delegate tx on GaiaAppGenState

	// create validator
	// TODO: wrap this in a function?
	cvStr := fmt.Sprintf("gaiacli tx stake create-validator %v", f.GaiaCliFlags())
	cvStr += fmt.Sprintf(" --from=%s", "bar")
	cvStr += fmt.Sprintf(" --pubkey=%s", consPubKey)
	cvStr += fmt.Sprintf(" --amount=%v", bondDenomTokens(2))
	cvStr += fmt.Sprintf(" --moniker=%v", "bar-vally")
	cvStr += fmt.Sprintf(" --commission-rate=%v", "0.05")
	cvStr += fmt.Sprintf(" --commission-max-rate=%v", "0.20")
	cvStr += fmt.Sprintf(" --commission-max-change-rate=%v", "0.10")

	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewDec(1))

	// Test --generate-only
	success, stdout, stderr := executeWriteRetStdStreams(t, cvStr+" --generate-only", app.DefaultKeyPass)
	require.True(t, success)
	require.Empty(t, stderr)
	msg := f.UnmarshalStdTx(stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success = executeWrite(t, cvStr+" --dry-run", app.DefaultKeyPass)
	require.True(t, success)

	executeWrite(t, cvStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, f.port)

	// Fetch account and ensure state matches desired
	barAcc = f.GetAccount(barAddr)
	require.Equal(t, int64(8), barAcc.GetCoins().AmountOf(bondDenom).Int64(), "%v", barAcc)

	// Fetch validator details and check them
	validator := f.executeGetValidator(t, sdk.ValAddress(barAddr))
	require.Equal(t, validator.OperatorAddr, sdk.ValAddress(barAddr))
	require.True(sdk.DecEq(t, sdk.NewDec(2), validator.Tokens))

	// Fetch delegations to the barAddr and test response
	validatorDelegations := f.executeGetValidatorDelegations(t, sdk.ValAddress(barAddr))
	require.Len(t, validatorDelegations, 1)
	require.NotZero(t, validatorDelegations[0].Shares)

	// unbond a single share
	success = f.executeUnbondValidator(t, keyBar, 1, sdk.ValAddress(barAddr))
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.port)

	// Ensure that validator has right amount of tokens
	validator = f.executeGetValidator(t, sdk.ValAddress(barAddr))
	require.Equal(t, "1.0000000000", validator.Tokens.String())

	// Check the unbonding delegations against expected state
	validatorUbds := f.executeGetValidatorUnbondingDelegations(t, sdk.ValAddress(barAddr))
	require.Len(t, validatorUbds, 1)
	require.Equal(t, "1", validatorUbds[0].Balance.Amount.String())

	// Check the staking parameters
	params := f.executeGetStakeParams(t)
	require.True(t, defaultParams.Equal(params))

	// Ensure that the staking pool is correct state
	pool := f.executeGetPool(t)
	require.Equal(t, initialPool.BondedTokens, pool.BondedTokens)

	// Cleanup on success
	f.CleanupDirs()
}

func TestGaiaCLISubmitProposal(t *testing.T) {
	// t.Parallel()
	f := InitFixtures(t)

	// Start gaiad server
	proc := f.StartGaiad(t)
	t.Log("Starting gaiad")
	defer proc.Stop(false)

	f.executeGetDepositParam(t)
	f.executeGetVotingParam(t)
	f.executeGetTallyingParam(t)

	// Get foo and ensure balance is correct
	fooAddr := f.GetKeyAddress(keyFoo)
	fooAcc := f.GetAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	proposalsQuery, _ := tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals %v", f.GaiaCliFlags()), "")
	require.Equal(t, "No matching proposals found", proposalsQuery)

	// submit a test proposal
	spStr := fmt.Sprintf("gaiacli tx gov submit-proposal %v", f.GaiaCliFlags())
	spStr += fmt.Sprintf(" --from=%s", "foo")
	spStr += fmt.Sprintf(" --deposit=%s", bondDenomTokens(5))
	spStr += fmt.Sprintf(" --type=%s", "Text")
	spStr += fmt.Sprintf(" --title=%s", "Test")
	spStr += fmt.Sprintf(" --description=%s", "test")

	// Test generate only
	success, stdout, stderr := executeWriteRetStdStreams(t, spStr+" --generate-only", app.DefaultKeyPass)
	require.True(t, success)
	require.True(t, success)
	require.Empty(t, stderr)
	msg := f.UnmarshalStdTx(stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success = executeWrite(t, spStr+" --dry-run", app.DefaultKeyPass)
	require.True(t, success)

	executeWrite(t, spStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, f.port)

	txs := f.executeGetTxs(t, "action:submit_proposal", fmt.Sprintf("proposer:%s", fooAddr))
	require.Len(t, txs, 1)

	// Ensure that deposit was taken out of foo
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(45), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Fetch proposal details and ensure they are correct
	proposal1 := f.executeGetProposal(t, 1)
	require.Equal(t, uint64(1), proposal1.GetProposalID())
	require.Equal(t, gov.StatusDepositPeriod, proposal1.GetStatus())

	// Ensure that proposal is active
	proposalsQuery, _ = tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals %v", f.GaiadFlags()), "")
	require.Equal(t, "  1 - Test", proposalsQuery)

	// Fetch amount deposited to the proposal and ensure correct
	deposit := f.executeGetDeposit(t, 1, fooAddr)
	require.Equal(t, int64(5), deposit.Amount.AmountOf(bondDenom).Int64())

	// Make a deposit to the proposal
	depositStr := fmt.Sprintf("gaiacli tx gov deposit --proposal-id 1 --deposit %s %v", fmt.Sprintf("10%s", bondDenom), f.GaiaCliFlags())
	depositStr += fmt.Sprintf(" --from=%s", "foo")

	// Test generate only
	success, stdout, stderr = executeWriteRetStdStreams(t, depositStr+" --generate-only", app.DefaultKeyPass)
	require.True(t, success)
	require.True(t, success)
	require.Empty(t, stderr)
	msg = f.UnmarshalStdTx(stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	executeWrite(t, depositStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, f.port)

	// Query the deposit and ensure that it is correct
	deposits := f.executeGetDeposits(t, 1)
	require.Len(t, deposits, 1)
	require.Equal(t, int64(15), deposits[0].Amount.AmountOf(bondDenom).Int64())

	// Get the deposit and ensure correct
	deposit = f.executeGetDeposit(t, 1, fooAddr)
	require.Equal(t, int64(15), deposit.Amount.AmountOf(bondDenom).Int64())

	// Ensure that transaction has appropriate tags
	txs = f.executeGetTxs(t, "action:deposit", fmt.Sprintf("depositor:%s", fooAddr))
	require.Len(t, txs, 1)

	// Ensure that foo has proper balance
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(35), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Ensure that proposal has entered the voting period
	proposal1 = f.executeGetProposal(t, 1)
	require.Equal(t, uint64(1), proposal1.GetProposalID())
	require.Equal(t, gov.StatusVotingPeriod, proposal1.GetStatus())

	voteStr := fmt.Sprintf("gaiacli tx gov vote --proposal-id 1 --option Yes %v", f.GaiaCliFlags())
	voteStr += fmt.Sprintf(" --from=%s", "foo")

	// Test generate only
	success, stdout, stderr = executeWriteRetStdStreams(t, voteStr+" --generate-only", app.DefaultKeyPass)
	require.True(t, success)
	require.Empty(t, stderr)
	msg = f.UnmarshalStdTx(stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	executeWrite(t, voteStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, f.port)

	// Fetch the vote details and ensure correct
	vote := f.executeGetVote(t, 1, fooAddr)
	require.Equal(t, uint64(1), vote.ProposalID)
	require.Equal(t, gov.OptionYes, vote.Option)

	// Fetch details for all votes and ensure they are correct
	votes := f.executeGetVotes(t, 1)
	require.Len(t, votes, 1)
	require.Equal(t, uint64(1), votes[0].ProposalID)
	require.Equal(t, gov.OptionYes, votes[0].Option)

	// Check the tags on voting transactions
	txs = f.executeGetTxs(t, "action:vote", fmt.Sprintf("voter:%s", fooAddr))
	require.Len(t, txs, 1)

	proposalsQuery, _ = tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals --status=DepositPeriod %v", f.GaiadFlags()), "")
	require.Equal(t, "No matching proposals found", proposalsQuery)

	proposalsQuery, _ = tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals --status=VotingPeriod %v", f.GaiadFlags()), "")
	require.Equal(t, "  1 - Test", proposalsQuery)

	// submit a second test proposal
	spStr = fmt.Sprintf("gaiacli tx gov submit-proposal %v", f.GaiaCliFlags())
	spStr += fmt.Sprintf(" --from=%s", "foo")
	spStr += fmt.Sprintf(" --deposit=%s", fmt.Sprintf("5%s", bondDenom))
	spStr += fmt.Sprintf(" --type=%s", "Text")
	spStr += fmt.Sprintf(" --title=%s", "Apples")
	spStr += fmt.Sprintf(" --description=%s", "test")

	executeWrite(t, spStr, app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, f.port)

	proposalsQuery, _ = tests.ExecuteT(t, fmt.Sprintf("gaiacli query gov proposals --limit=1 %v", f.GaiaCliFlags()), "")
	require.Equal(t, "  2 - Apples", proposalsQuery)

	// Cleanup directories on success
	f.CleanupDirs()
}

func TestGaiaCLISendGenerateSignAndBroadcast(t *testing.T) {
	// t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.StartGaiad(t)
	t.Log("Starting gaiad")
	defer proc.Stop(false)
	cdc := app.MakeCodec()

	// Pull the two addresses
	fooAddr := f.GetKeyAddress(keyFoo)
	barAddr := f.GetKeyAddress(keyBar)

	// Test generate sendTx with default gas
	response := f.sendCoinsResponse(t, bondDenomTokens(10), keyFoo, barAddr, "--generate-only")
	require.Equal(t, response.Fee.Gas, uint64(client.DefaultGasLimit))
	require.Equal(t, len(response.Msgs), 1)
	require.Equal(t, 0, len(response.GetSignatures()))

	// Test generate sendTx with default gas
	response = f.sendCoinsResponse(t, bondDenomTokens(10), keyFoo, barAddr, "--gas=100", "--generate-only")
	require.Equal(t, response.Fee.Gas, uint64(100))
	require.Equal(t, len(response.Msgs), 1)
	require.Equal(t, 0, len(response.GetSignatures()))

	// Test generate sendTx with default gas
	response = f.sendCoinsResponse(t, bondDenomTokens(10), keyFoo, barAddr, "--gas=simulate", "--generate-only")
	require.True(t, response.Fee.Gas > 0)
	require.Equal(t, len(response.Msgs), 1)

	// Marshall the response and save to file
	res, err := cdc.MarshalJSON(response)
	require.NoError(t, err)

	// Write the output to disk
	unsignedTxFile := writeToNewTempFile(t, string(res))
	defer os.Remove(unsignedTxFile.Name())

	// Test sign --validate-signatures
	success, stdout, _ := executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx sign %v --validate-signatures %v", f.GaiaCliFlags(), unsignedTxFile.Name()))
	require.False(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n 0: %v\n\nSignatures:\n\n", fooAddr.String()), stdout)

	// Test sign
	success, stdout, _ = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx sign %v --name=foo %v", f.GaiaCliFlags(), unsignedTxFile.Name()), app.DefaultKeyPass)
	require.True(t, success)
	msg := f.UnmarshalStdTx(stdout)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 1, len(msg.GetSignatures()))
	require.Equal(t, fooAddr.String(), msg.GetSigners()[0].String())

	// Write the output to disk
	signedTxFile := writeToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Test sign --print-signatures
	success, stdout, _ = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx sign %v --validate-signatures %v", f.GaiaCliFlags(), signedTxFile.Name()))
	require.True(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n 0: %v\n\nSignatures:\n 0: %v\t[OK]\n\n", fooAddr.String(),
		fooAddr.String()), stdout)

	// Test broadcast
	fooAcc := f.GetAccount(fooAddr)
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	success, stdout, _ = executeWriteRetStdStreams(t, fmt.Sprintf(
		"gaiacli tx broadcast %v --json %v", f.GaiaCliFlags(), signedTxFile.Name()))
	require.True(t, success)

	result := DeliverTxResponse{}
	require.Nil(t, cdc.UnmarshalJSON([]byte(stdout), &result))
	require.Equal(t, msg.Fee.Gas, uint64(result.Response.GasUsed))
	require.Equal(t, msg.Fee.Gas, uint64(result.Response.GasWanted))
	tests.WaitForNextNBlocksTM(1, f.port)

	// Ensure account balances are correct
	barAcc := f.GetAccount(barAddr)
	fooAcc = f.GetAccount(fooAddr)
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(bondDenom).Int64())
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(bondDenom).Int64())

	// Cleanup on success
	f.CleanupDirs()
}

func TestGaiaCLIConfig(t *testing.T) {
	// t.Parallel()
	f := InitFixtures(t)
	node := fmt.Sprintf("%s:%s", f.rpcAddr, f.port)
	executeWrite(t, fmt.Sprintf("gaiacli --home=%s config", f.gaiadHome), f.gaiacliHome, node, "y")
	config, err := ioutil.ReadFile(path.Join(f.gaiacliHome, "config", "config.toml"))
	require.NoError(t, err)
	expectedConfig := fmt.Sprintf(`chain_id = "%s"
home = "%s"
node = "%s"
output = "text"
trace = false
trust_node = true
`, f.chainID, f.gaiacliHome, node)
	require.Equal(t, expectedConfig, string(config))
	// ensure a backup gets created
	executeWrite(t, "gaiacli config", f.gaiacliHome, node, "y", "y")
	configBackup, err := ioutil.ReadFile(path.Join(f.gaiacliHome, "config", "config.toml-old"))
	require.NoError(t, err)
	require.Equal(t, expectedConfig, string(configBackup))

	require.NoError(t, os.RemoveAll(f.gaiadHome))
	executeWrite(t, "gaiacli config", f.gaiacliHome, node, "y")

	// ensure it works without an initialized gaiad state
	expectedConfig = fmt.Sprintf(`chain_id = ""
home = "%s"
node = "%s"
output = "text"
trace = false
trust_node = true
`, f.gaiacliHome, node)
	config, err = ioutil.ReadFile(path.Join(f.gaiacliHome, "config", "config.toml"))
	require.NoError(t, err)
	require.Equal(t, expectedConfig, string(config))

	// Cleanup on success
	f.CleanupDirs()
}
