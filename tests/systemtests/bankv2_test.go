//go:build system_test

package systemtests

import (
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestBankV2SendTxCmd(t *testing.T) {
	// Currently only run with app v2
	if !isV2() {
		t.Skip()
	}
	// scenario: test bank send command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	sut.StartChain(t)

	// query validator balance and make sure it has enough balance
	var transferAmount int64 = 1000
	raw := cli.CustomQuery("q", "bankv2", "balance", valAddr, denom)
	valBalance := gjson.Get(raw, "balance.amount").Int()

	require.Greater(t, valBalance, transferAmount, "not enough balance found with validator")

	bankSendCmdArgs := []string{"tx", "bankv2", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom)}

	// test valid transaction
	rsp := cli.Run(append(bankSendCmdArgs, "--fees=1stake")...)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

	// Check balance after send
	valRaw := cli.CustomQuery("q", "bankv2", "balance", valAddr, denom)
	valBalanceAfer := gjson.Get(valRaw, "balance.amount").Int()

	// TODO: Make DeductFee ante handler work with bank/v2
	require.Equal(t, valBalanceAfer, valBalance-transferAmount)

	receiverRaw := cli.CustomQuery("q", "bankv2", "balance", receiverAddr, denom)
	receiverBalance := gjson.Get(receiverRaw, "balance.amount").Int()
	require.Equal(t, receiverBalance, transferAmount)
}

func TestCreateDenom(t *testing.T) {
	// Currently only run with app v2
	if !isV2() {
		t.Skip()
	}
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// add new key
	denom := "stake"
	subDenom := "test"
	feeAmount := math.NewInt(1000000)

	sut.ModifyGenesisJSON(
		t,
		SetDenomCreationFee(t, denom, feeAmount),
	)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, valAddr)

	sut.StartChain(t)

	raw := cli.CustomQuery("q", "bankv2", "balance", valAddr, denom)
	valBalanceBefore := gjson.Get(raw, "balance.amount").Int()

	rsp := cli.Run("tx", "bankv2", "create-denom", subDenom, "--from", valAddr)
	txResult, found := cli.AwaitTxCommitted(rsp)
	fmt.Println("txResult", txResult)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

	raw = cli.CustomQuery("q", "bankv2", "balance", valAddr, denom)
	valBalanceAfter := gjson.Get(raw, "balance.amount").Int()

	require.Equal(t, valBalanceBefore-valBalanceAfter, feeAmount.Int64())

	raw = cli.CustomQuery("q", "bankv2", "denoms-from-creator", valAddr)
	newDenoms := gjson.Get(raw, "denoms").Array()
	require.Equal(t, len(newDenoms), 1)

	raw = cli.CustomQuery("q", "bankv2", "denom-authority-metadata", newDenoms[0].String())
	admin := gjson.Get(raw, "authority_metadata.admin").String()
	require.Equal(t, admin, valAddr)
}

func TestMintBurnTokenCmd(t *testing.T) {
	// Currently only run with app v2
	if !isV2() {
		t.Skip()
	}
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// add new key
	denom := "stake"
	subDenom := "test"
	feeAmount := math.NewInt(1_000_000)
	mintAmount := 1_000_000
	burnAmount := 500_000

	sut.ModifyGenesisJSON(
		t,
		SetDenomCreationFee(t, denom, feeAmount),
	)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")

	sut.StartChain(t)

	rsp := cli.Run("tx", "bankv2", "create-denom", subDenom, "--from", valAddr)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

	raw := cli.CustomQuery("q", "bankv2", "denoms-from-creator", valAddr)
	newDenoms := gjson.Get(raw, "denoms").Array()
	require.Equal(t, len(newDenoms), 1)

	// non tokenfactory should not be minted directly
	rsp = cli.Run("tx", "bankv2", "mint", valAddr, receiverAddr, fmt.Sprintf("%d%s", mintAmount, denom))
	code := gjson.Get(rsp, "code").Int()
	require.NotEqual(t, code, int64(0))
	rawLog := gjson.Get(rsp, "raw_log").String()
	require.Contains(t, rawLog, "invalid authority")

	rsp = cli.Run("tx", "bankv2", "mint", valAddr, receiverAddr, fmt.Sprintf("%d%s", mintAmount, newDenoms[0]))
	txResult, found = cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

	raw = cli.CustomQuery("q", "bankv2", "balance", receiverAddr, newDenoms[0].String())
	balance := gjson.Get(raw, "balance.amount").Int()
	require.Equal(t, balance, int64(mintAmount))

	// Burn token from receiver
	rsp = cli.Run("tx", "bankv2", "burn", valAddr, receiverAddr, fmt.Sprintf("%d%s", burnAmount, newDenoms[0]))
	txResult, found = cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

	raw = cli.CustomQuery("q", "bankv2", "balance", receiverAddr, newDenoms[0].String())
	balance = gjson.Get(raw, "balance.amount").Int()
	require.Equal(t, balance, int64(mintAmount-burnAmount))
}

func TestMintBurnTokenProposal(t *testing.T) {
	// Currently only run with app v2
	if !isV2() {
		t.Skip()
	}
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// add new key
	denom := "stake"
	subDenom := "test"
	mintAmount := 1_000_000
	burnAmount := 500_000

	sut.ModifyGenesisJSON(
		t,
		SetGovVotingPeriod(t, time.Second*8),
		SetGovExpeditedVotingPeriod(t, time.Second*7),
	)

	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	valAddr1 := cli.GetKeyAddr("node1")
	require.NotEmpty(t, valAddr1)

	// add new key
	receiverAddr := cli.AddKey("account1")

	sut.StartChain(t)

	// get gov module address
	resp := cli.CustomQuery("q", "auth", "module-account", "gov")
	govAddr := gjson.Get(resp, "account.value.address").String()

	rsp := cli.Run("tx", "bankv2", "create-denom", subDenom, "--from", valAddr)
	txResult, found := cli.AwaitTxCommitted(rsp)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

	raw := cli.CustomQuery("q", "bankv2", "denoms-from-creator", valAddr)
	newDenoms := gjson.Get(raw, "denoms").Array()
	require.Equal(t, len(newDenoms), 1)
	fmt.Println("denoms", raw)

	invalidMintProposal := fmt.Sprintf(`
	{
 "messages": [
  {
   "@type": "/cosmos.bank.v2.MsgMint",
   "authority": "%s",
   "to_address": "%s",
   "amount": {
     "denom": "%s",
     "amount": "%d"
   }
  }
 ],
 "metadata": "ipfs://CID",
 "deposit": "100000000stake",
 "title": "mint tokenfactory token",
 "summary": "testing"
}`, govAddr, receiverAddr, newDenoms[0], mintAmount)

	invalidMintProposalFile := StoreTempFile(t, []byte(invalidMintProposal))
	rsp = cli.RunAndWait("tx", "gov", "submit-proposal", invalidMintProposalFile.Name(), "--from", valAddr)
	RequireTxSuccess(t, rsp)

	// vote to proposal from two validators
	rsp = cli.RunAndWait("tx", "gov", "vote", "1", "yes", "--from", valAddr)
	RequireTxSuccess(t, rsp)
	rsp = cli.RunAndWait("tx", "gov", "vote", "1", "yes", "--from", valAddr1)
	RequireTxSuccess(t, rsp)

	time.Sleep(8 * time.Second)

	// Token should not be minted to receiver
	raw = cli.CustomQuery("q", "bankv2", "balance", receiverAddr, newDenoms[0].String())
	balance := gjson.Get(raw, "balance.amount").Int()
	require.Equal(t, balance, int64(0))

	validMintProposal := fmt.Sprintf(`
	{
 "messages": [
  {
   "@type": "/cosmos.bank.v2.MsgMint",
   "authority": "%s",
   "to_address": "%s",
   "amount": {
     "denom": "%s",
     "amount": "%d"
   }
  }
 ],
 "metadata": "ipfs://CID",
 "deposit": "100000000stake",
 "title": "mint tokenfactory token",
 "summary": "testing"
}`, govAddr, receiverAddr, denom, mintAmount)

	mintProposalFile := StoreTempFile(t, []byte(validMintProposal))

	rsp = cli.RunAndWait("tx", "gov", "submit-proposal", mintProposalFile.Name(), "--from", valAddr)
	RequireTxSuccess(t, rsp)

	// vote to proposal from two validators
	rsp = cli.RunAndWait("tx", "gov", "vote", "2", "yes", "--from", valAddr)
	RequireTxSuccess(t, rsp)
	rsp = cli.RunAndWait("tx", "gov", "vote", "2", "yes", "--from", valAddr1)
	RequireTxSuccess(t, rsp)

	time.Sleep(8 * time.Second)

	// stake should be minted to receiver
	raw = cli.CustomQuery("q", "bankv2", "balance", receiverAddr, denom)
	balance = gjson.Get(raw, "balance.amount").Int()
	require.Equal(t, balance, int64(mintAmount))

	validBurnProposal := fmt.Sprintf(`
	{
 "messages": [
  {
   "@type": "/cosmos.bank.v2.MsgBurn",
   "authority": "%s",
   "burn_from_address": "%s",
   "amount": {
     "denom": "%s",
     "amount": "%d"
   }
  }
 ],
 "metadata": "ipfs://CID",
 "deposit": "100000000stake",
 "title": "mint tokenfactory token",
 "summary": "testing"
}`, govAddr, receiverAddr, denom, burnAmount)

	burnProposalFile := StoreTempFile(t, []byte(validBurnProposal))

	rsp = cli.RunAndWait("tx", "gov", "submit-proposal", burnProposalFile.Name(), "--from", valAddr)
	RequireTxSuccess(t, rsp)

	// vote to proposal from two validators
	rsp = cli.RunAndWait("tx", "gov", "vote", "3", "yes", "--from", valAddr)
	RequireTxSuccess(t, rsp)
	rsp = cli.RunAndWait("tx", "gov", "vote", "3", "yes", "--from", valAddr1)
	RequireTxSuccess(t, rsp)

	time.Sleep(8 * time.Second)

	// stake should be burned from receiver
	raw = cli.CustomQuery("q", "bankv2", "balance", receiverAddr, denom)
	balance = gjson.Get(raw, "balance.amount").Int()
	require.Equal(t, balance, int64(mintAmount-burnAmount))
}
