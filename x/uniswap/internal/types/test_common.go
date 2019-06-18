package types

import (
	"github.com/tendermint/tendermint/crypto/ed25519"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	amt = sdk.NewInt(100)

	sender_pk    = ed25519.GenPrivKey().PubKey()
	recipient_pk = ed25519.GenPrivKey().PubKey()
	sender       = sdk.AccAddress(sender_pk.Address())
	recipient    = sdk.AccAddress(recipient_pk.Address())

	denom0 = "atom"
	denom1 = "btc"

	input    = sdk.NewCoin(denom0, sdk.NewInt(1000))
	output   = sdk.NewCoin(denom1, sdk.NewInt(500))
	deadline = time.Now()

	emptyAddr  sdk.AccAddress
	emptyDenom = "   "
	emptyTime  time.Time
)
