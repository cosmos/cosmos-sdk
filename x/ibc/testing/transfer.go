package testing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewTransferCoins(dst TestChannel, denom string, amount int64) sdk.Coins {
	return sdk.NewCoins(sdk.NewCoin(fmt.Sprintf("%s/%s/%s", dst.PortID, dst.ID, denom), sdk.NewInt(amount)))
}
