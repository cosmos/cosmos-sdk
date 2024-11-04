//go:build system_test

package systemtests

import (
	"fmt"
	"testing"

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
		SetDenomCreationFee(t, "stake", feeAmount),
	)

	// get validator address
	valAddr := gjson.Get(cli.Keys("keys", "list"), "1.address").String()
	require.NotEmpty(t, valAddr)

	sut.StartChain(t)

	raw := cli.CustomQuery("q", "bankv2", "balance", valAddr, denom)
	valBalanceBefore := gjson.Get(raw, "balance.amount").Int()
	fmt.Println("valBalance", valBalanceBefore)

	rsp := cli.Run("tx", "bankv2", "create-denom", subDenom, "--from", valAddr)
	txResult, found := cli.AwaitTxCommitted(rsp)
	fmt.Println("txResult", txResult)
	require.True(t, found)
	RequireTxSuccess(t, txResult)

	raw = cli.CustomQuery("q", "bankv2", "balance", valAddr, denom)
	valBalanceAfter := gjson.Get(raw, "balance.amount").Int()
	fmt.Println("valBalanceAfter", valBalanceAfter)

	require.Equal(t, valBalanceBefore - valBalanceAfter, feeAmount.Int64())
}
