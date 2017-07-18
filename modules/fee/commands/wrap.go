package commands

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin"
	bcmd "github.com/tendermint/basecoin/cmd/basecli/commands"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
)

//nolint
const (
	FlagFee = "fee"
)

// FeeWrapper wraps a tx with an optional fee payment
type FeeWrapper struct{}

var _ bcmd.Wrapper = FeeWrapper{}

// Wrap checks for FlagFee and if present wraps the tx with a
// FeeTx of the given amount, paid by the signer
func (FeeWrapper) Wrap(tx basecoin.Tx) (res basecoin.Tx, err error) {
	//parse the fee and amounts into coin types
	toll, err := coin.ParseCoin(viper.GetString(FlagFee))
	if err != nil {
		return res, err
	}
	// if no fee, do nothing, otherwise wrap it
	if toll.IsZero() {
		return tx, nil
	}
	res = fee.NewFee(tx, toll, bcmd.GetSignerAct())
	return
}

// Register adds the sequence flags to the cli
func (FeeWrapper) Register(fs *pflag.FlagSet) {
	fs.String(FlagFee, "0mycoin", "Coins for the transaction fee of the format <amt><coin>")
}
