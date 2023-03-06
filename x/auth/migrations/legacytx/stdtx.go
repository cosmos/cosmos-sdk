package legacytx

import (
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Interface implementation checks
var (
	_ codectypes.UnpackInterfacesMessage = (*StdSignature)(nil)
)

// StdFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
// [Deprecated]
type StdFee struct {
	Amount  sdk.Coins `json:"amount" yaml:"amount"`
	Gas     uint64    `json:"gas" yaml:"gas"`
	Payer   string    `json:"payer,omitempty" yaml:"payer"`
	Granter string    `json:"granter,omitempty" yaml:"granter"`
}

// Deprecated: NewStdFee returns a new instance of StdFee
func NewStdFee(gas uint64, amount sdk.Coins) StdFee {
	return StdFee{
		Amount: amount,
		Gas:    gas,
	}
}

// GetGas returns the fee's (wanted) gas.
func (fee StdFee) GetGas() uint64 {
	return fee.Gas
}

// GetAmount returns the fee's amount.
func (fee StdFee) GetAmount() sdk.Coins {
	return fee.Amount
}

// Bytes returns the encoded bytes of a StdFee.
func (fee StdFee) Bytes() []byte {
	if len(fee.Amount) == 0 {
		fee.Amount = sdk.NewCoins()
	}

	bz, err := legacy.Cdc.MarshalJSON(fee)
	if err != nil {
		panic(err)
	}

	return bz
}

// GasPrices returns the gas prices for a StdFee.
//
// NOTE: The gas prices returned are not the true gas prices that were
// originally part of the submitted transaction because the fee is computed
// as fee = ceil(gasWanted * gasPrices).
func (fee StdFee) GasPrices() sdk.DecCoins {
	return sdk.NewDecCoinsFromCoins(fee.Amount...).QuoDec(math.LegacyNewDec(int64(fee.Gas)))
}

// StdTip is the tips used in a tipped transaction.
type StdTip struct {
	Amount sdk.Coins `json:"amount" yaml:"amount"`
	Tipper string    `json:"tipper" yaml:"tipper"`
}
