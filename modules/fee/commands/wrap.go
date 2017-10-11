package commands

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
)

//nolint
const (
	FlagFee   = "fee"
	FlagPayer = "payer"
)

// FeeWrapper wraps a tx with an optional fee payment
type FeeWrapper struct{}

var _ txcmd.Wrapper = FeeWrapper{}

// Wrap checks for FlagFee and if present wraps the tx with a
// FeeTx of the given amount, paid by the signer
func (FeeWrapper) Wrap(tx sdk.Tx) (res sdk.Tx, err error) {
	//parse the fee and amounts into coin types
	toll, err := coin.ParseCoin(viper.GetString(FlagFee))
	if err != nil {
		return res, err
	}
	// if no fee, do nothing, otherwise wrap it
	if toll.IsZero() {
		return tx, nil
	}

	payer, err := readPayer()
	if err != nil {
		return res, err
	}

	res = fee.NewFee(tx, toll, payer)
	return
}

// Register adds the sequence flags to the cli
func (FeeWrapper) Register(fs *pflag.FlagSet) {
	fs.String(FlagFee, "0mycoin", "Coins for the transaction fee of the format <amt><coin>")
	fs.String(FlagPayer, "", "Account to pay fee if not current signer (for multisig)")
}

func readPayer() (sdk.Actor, error) {
	payer := viper.GetString(FlagPayer)
	if payer == "" {
		return txcmd.GetSignerAct(), nil
	}
	return commands.ParseActor(payer)
}
