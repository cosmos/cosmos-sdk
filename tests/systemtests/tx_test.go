//go:build system_test

package systemtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryBySig(t *testing.T) {
	sut.ResetChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new key
	receiverAddr := cli.AddKey("account1")
	denom := "stake"
	var transferAmount int64 = 1000

	sut.StartChain(t)

	// qc := tx.NewServiceClient(sut.RPCClient(t))

	// create unsign tx
	bankSendCmdArgs := []string{"tx", "bank", "send", valAddr, receiverAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := StoreTempFile(t, []byte(res))
	fmt.Println("txFile", txFile)

	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), fmt.Sprintf("--from=%s", valAddr), fmt.Sprintf("--chain-id=%s",sut.chainID), fmt.Sprintf("--output-document=%s", txFile.Name()))
	fmt.Println("res", res)
}
