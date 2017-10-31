package commands

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/modules/auth"
)

//nolint
const (
	FlagMulti = "multi"
)

// SigWrapper wraps a tx with a signature layer to hold pubkey sigs
type SigWrapper struct{}

var _ txcmd.Wrapper = SigWrapper{}

// Wrap will wrap the tx with OneSig or MultiSig depending on flags
func (SigWrapper) Wrap(tx sdk.Tx) (res sdk.Tx, err error) {
	if !viper.GetBool(FlagMulti) {
		res = auth.NewSig(tx).Wrap()
	} else {
		res = auth.NewMulti(tx).Wrap()
	}
	return
}

// Register adds the sequence flags to the cli
func (SigWrapper) Register(fs *pflag.FlagSet) {
	fs.Bool(FlagMulti, false, "Prepare the tx for multisig")
}
