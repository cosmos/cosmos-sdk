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
	proc := f.GDStart(fmt.Sprintf("--minimum_fees=%s", sdk.NewInt64Coin(feeDenom, 2)))
	defer proc.Stop(false)

	// Check the amount of coins in the foo account to ensure that the right amount exists
	fooAcc := f.QueryAccount(f.KeyAddress(keyFoo))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	// Send a transaction that will get rejected
	success := f.TxSend(keyFoo, f.KeyAddress(keyBar), sdk.NewInt64Coin(denom, 10))
	require.False(f.T, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestGaiaCLIFeesDeduction(t *testing.T) {
	t.Parallel()
	f := initializeFixtures(t)
	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", f.GCLIHome, f.RPCAddr, f.ChainID)

	// start gaiad server with minimum fees
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("gaiad start --home=%s --rpc.laddr=%v --p2p.laddr=%v --minimum_fees=1footoken", f.GDHome, f.RPCAddr, f.P2PAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(f.Port)
	tests.WaitForNextNBlocksTM(1, f.Port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show foo --home=%s", f.GCLIHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("gaiacli keys show bar --home=%s", f.GCLIHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf("footoken").Int64())

	// test simulation
	success := executeWrite(t, fmt.Sprintf(
		"gaiacli tx send %v --amount=1000footoken --to=%s --from=foo --fees=1footoken --dry-run", flags, barAddr), app.DefaultKeyPass)
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)
	// ensure state didn't change
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf("footoken").Int64())

	// insufficient funds (coins + fees)
	success = executeWrite(t, fmt.Sprintf(
		"gaiacli tx send %v --amount=1000footoken --to=%s --from=foo --fees=1footoken", flags, barAddr), app.DefaultKeyPass)
	require.False(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)
	// ensure state didn't change
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(1000), fooAcc.GetCoins().AmountOf("footoken").Int64())

	// test success (transfer = coins + fees)
	success = executeWrite(t, fmt.Sprintf(
		"gaiacli tx send %v --fees=300footoken --amount=500footoken --to=%s --from=foo", flags, barAddr), app.DefaultKeyPass)
	require.True(t, success)
	cleanupDirs(f.GDHome, f.GCLIHome)
}

func TestGaiaCLISend(t *testing.T) {
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

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	executeWrite(t, fmt.Sprintf("gaiacli tx send %v --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, f.Port)

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
	tests.WaitForNextNBlocksTM(1, f.Port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(20), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(30), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	// test memo
	executeWrite(t, fmt.Sprintf("gaiacli tx send %v --amount=10%s --to=%s --from=foo --memo 'testmemo'", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, f.Port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(30), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(20), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	cleanupDirs(f.GDHome, f.GCLIHome)
}

func TestGaiaCLIGasAuto(t *testing.T) {
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

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(50), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	// Test failure with auto gas disabled and very little gas set by hand
	success := executeWrite(t, fmt.Sprintf("gaiacli tx send %v --gas=10 --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	require.False(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)
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
	success, stdout, _ := executeWriteRetStdStreams(t, fmt.Sprintf("gaiacli tx send %v --json --gas=auto --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
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
	tests.WaitForNextNBlocksTM(1, f.Port)
	// Check state has changed accordingly
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	cleanupDirs(f.GDHome, f.GCLIHome)
}

func TestGaiaCLICreateValidator(t *testing.T) {
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
	consPubKey := sdk.MustBech32ifyConsPub(ed25519.GenPrivKey().PubKey())

	executeWrite(t, fmt.Sprintf("gaiacli tx send %v --amount=10%s --to=%s --from=foo", flags, stakeTypes.DefaultBondDenom, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(1, f.Port)

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(10), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())
	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", fooAddr, flags))
	require.Equal(t, int64(40), fooAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64())

	defaultParams := stake.DefaultParams()
	initialPool := stake.InitialPool()
	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewInt(100)) // Delegate tx on GaiaAppGenState

	// create validator
	cvStr := fmt.Sprintf("gaiacli tx stake create-validator %v", flags)
	cvStr += fmt.Sprintf(" --from=%s", "bar")
	cvStr += fmt.Sprintf(" --pubkey=%s", consPubKey)
	cvStr += fmt.Sprintf(" --amount=%v", fmt.Sprintf("2%s", stakeTypes.DefaultBondDenom))
	cvStr += fmt.Sprintf(" --moniker=%v", "bar-vally")
	cvStr += fmt.Sprintf(" --commission-rate=%v", "0.05")
	cvStr += fmt.Sprintf(" --commission-max-rate=%v", "0.20")
	cvStr += fmt.Sprintf(" --commission-max-change-rate=%v", "0.10")

	initialPool.BondedTokens = initialPool.BondedTokens.Add(sdk.NewInt(1))

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
	tests.WaitForNextNBlocksTM(1, f.Port)

	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %s %v", barAddr, flags))
	require.Equal(t, int64(8), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64(), "%v", barAcc)

	validator := executeGetValidator(t, fmt.Sprintf("gaiacli query stake validator %s %v", sdk.ValAddress(barAddr), flags))
	require.Equal(t, validator.OperatorAddr, sdk.ValAddress(barAddr))
	require.True(sdk.IntEq(t, sdk.NewInt(2), validator.Tokens))

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
	tests.WaitForNextNBlocksTM(1, f.Port)

	/* // this won't be what we expect because we've only started unbonding, haven't completed
	barAcc = executeGetAccount(t, fmt.Sprintf("gaiacli query account %v %v", barCech, flags))
	require.Equal(t, int64(9), barAcc.GetCoins().AmountOf(stakeTypes.DefaultBondDenom).Int64(), "%v", barAcc)
	*/
	validator = executeGetValidator(t, fmt.Sprintf("gaiacli query stake validator %s %v", sdk.ValAddress(barAddr), flags))
	require.Equal(t, "1", validator.Tokens.String())

	validatorUbds := executeGetValidatorUnbondingDelegations(t,
		fmt.Sprintf("gaiacli query stake unbonding-delegations-from %s %v", sdk.ValAddress(barAddr), flags))
	require.Len(t, validatorUbds, 1)
	require.Equal(t, "1", validatorUbds[0].Balance.Amount.String())

	params := executeGetParams(t, fmt.Sprintf("gaiacli query stake parameters %v", flags))
	require.True(t, defaultParams.Equal(params))

	pool := executeGetPool(t, fmt.Sprintf("gaiacli query stake pool %v", flags))
	require.Equal(t, initialPool.BondedTokens, pool.BondedTokens)
	cleanupDirs(f.GDHome, f.GCLIHome)
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
