package types

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CreateDenom(typ, id string) string {
	nm := fmt.Sprintf("%s-%s", typ, id)
	nmHex := hex.EncodeToString([]byte(nm))
	denom := fmt.Sprintf("%s/%s", ModuleName, nmHex)
	return denom
}

func GetTypeAndIDFrom(coinDenom string) (typ, id string, err error) {
	prefix := fmt.Sprintf("%s/", ModuleName)
	if !strings.HasPrefix(coinDenom, prefix) {
		return typ, id, fmt.Errorf("invalid ntf denom: %s", coinDenom)
	}

	nmHex, err := hex.DecodeString(strings.TrimPrefix(coinDenom, prefix))
	if err != nil {
		return typ, id, fmt.Errorf("invalid ntf denom: %s", coinDenom)
	}

	result := strings.Split(string(nmHex), "/")
	if len(result) != 2 {
		return typ, id, fmt.Errorf("invalid ntf denom: %s", coinDenom)
	}
	return result[0], result[1], nil
}

func (n NFT) Coin() sdk.Coin {
	return sdk.NewCoin(CreateDenom(n.Type, n.ID), sdk.OneInt())
}
