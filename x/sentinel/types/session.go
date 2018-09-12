package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

var pool = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Session struct {
	TotalLockedCoins sdk.Coins
	ReleasedCoins    sdk.Coins
	Counter          int64
	Timestamp        int64
	VpnPubKey        crypto.PubKey
	CPubKey          crypto.PubKey
	CAddress         sdk.AccAddress
	Status           uint8
}

func GetNewSessionMap(coins sdk.Coins, vpnpub crypto.PubKey, cpub crypto.PubKey, caddr sdk.AccAddress, time int64) Session {
	return Session{
		TotalLockedCoins: coins,
		ReleasedCoins:    coins.Minus(coins),
		VpnPubKey:        vpnpub,
		CPubKey:          cpub,
		Timestamp:        time,
		CAddress:         caddr,
		Status:           1,
	}

}

/*Status of the Session
1 : for active session
0: for closed session


*/
