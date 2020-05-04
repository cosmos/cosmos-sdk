package testutil

import (
	"fmt"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxSign is simcli sign
func TxSign(f *cli.Fixtures, signer, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx sign %v --keyring-backend=test --from=%s %v", f.SimcliBinary, f.Flags(), signer, fileName)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxSend is simcli tx send
func TxSend(f *cli.Fixtures, from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx send --keyring-backend=test %s %s %s %v", f.SimcliBinary, from, to, amount, f.Flags())

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxValidateSignatures is simcli tx validate-signatures
func TxValidateSignatures(f *cli.Fixtures, fileName string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx validate-signatures %v --keyring-backend=test %v", f.SimcliBinary,
		f.Flags(), fileName)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}
