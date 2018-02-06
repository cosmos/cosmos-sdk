package coins

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
)

//-----------------------------------------------------------
// BaseAccount

var _ sdk.Account = (*BaseAccount)(nil)

// BaseAccount - base account structure.
// Extend this by embedding this in your AppAccount.
// See the examples/basecoin/types/account.go for an example.
type CoinAccount struct {
	Account  *auth.BaseAccount  `json:"account"`
	Coins    Coins              `json:"coins"`
}

func NewCoinAccountWithAccount(acc auth.Account) CoinAccount {
	return CoinAccount{
		Account: acc,
		Coins: NewCoins()
	}
}

func (acc *CoinAccount) GetCoins() Coins {
	return acc.Coins
}

func (acc *CoinAccount) SetCoins(coins Coins) {
	acc.Coins = coins
}

func (acc *CoinAccount) AddCoins(coins Coins) {
	acc.SetCoins(acc.GetCoins().Plus(coins))
}

// Implements sdk.Account.
func (acc *CoinAccount) SubtractCoins(coins Coins) {
	acc.SetCoins(acc.GetCoins().Minus(coins))
}


//----------------------------------------
// Wire

func RegisterWireCoinAccount(cdc *wire.Codec) {
	// Register crypto.[PubKey,PrivKey,Signature] types.
	crypto.RegisterWire(cdc)
}
