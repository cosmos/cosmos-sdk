package covenant

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Covenant struct {
	Settlers  []sdk.Address
	Receivers []sdk.Address
	Amount    sdk.Coins
}
