package types

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	zero      = sdk.NewInt(0)
	one       = sdk.NewInt(1)
	baseValue = sdk.NewInt(100)

	sender_pk    = ed25519.GenPrivKey().PubKey()
	recipient_pk = ed25519.GenPrivKey().PubKey()
	sender       = sdk.AccAddress(sender_pk.Address())
	recipient    = sdk.AccAddress(recipient_pk.Address())

	denom0 = "atom"
	denom1 = "btc"
	denom2 = "eth"

	coin     = sdk.NewCoin(denom1, sdk.NewInt(1000))
	amount   = sdk.NewCoins(coin)
	bound    = sdk.NewInt(100)
	deadline = time.Now()

	emptyAddr  sdk.AccAddress
	emptyDenom = "   "
	emptyTime  time.Time
)
