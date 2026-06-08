package aminojsonpb

import v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"

// FeeAmount is a simple denom/amount string pair used to build AminoSignFee
// without exposing cosmossdk.io/api/cosmos/base/v1beta1 to callers.
type FeeAmount struct {
	Denom  string
	Amount string
}

// NewAminoSignFee constructs an AminoSignFee from plain coin data, hiding
// the basev1beta1.Coin construction inside the internal aminojsonpb package.
func NewAminoSignFee(coins []FeeAmount, gas uint64, payer, granter string) *AminoSignFee {
	feeCoins := make([]*v1beta1.Coin, len(coins))
	for i, c := range coins {
		feeCoins[i] = &v1beta1.Coin{Denom: c.Denom, Amount: c.Amount}
	}
	return &AminoSignFee{Amount: feeCoins, Gas: gas, Payer: payer, Granter: granter}
}
