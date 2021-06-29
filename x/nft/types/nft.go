package types

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (n NFT) Coin() sdk.Coin {
	nm := fmt.Sprintf("%s-%s", n.Type, n.ID)
	nmHex := hex.EncodeToString([]byte(nm))
	denom := fmt.Sprintf("%s/%s", ModuleName, nmHex)
	return sdk.NewCoin(denom, sdk.OneInt())
}
