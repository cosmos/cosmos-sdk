package types

import (
	"math/rand"
	"time"

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

func GetNewSessionId() []byte {

	bytes := make([]byte, 20)
	for i := 0; i < 20; i++ {
		bytes[i] = pool[rand.Intn(len(pool))]
	}
	// fmt.Println(bytes)
	return bytes

}
func GetNewSessionMap(coins sdk.Coins, vpnpub crypto.PubKey, cpub crypto.PubKey, caddr sdk.AccAddress) Session {
	ti := time.Now().UnixNano()
	return Session{
		TotalLockedCoins: coins,
		ReleasedCoins:    coins.Minus(coins),
		VpnPubKey:        vpnpub,
		CPubKey:          cpub,
		Timestamp:        ti,
		CAddress:         caddr,
		Status:           1,
	}

}

/*Status of the Session
1 : for active session
0: for closed session


*/
