package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// DenomManager specifies rules for minting, sending and burning coins.
type DenomManager interface {
	// OnMint specifies a rule for minting coins and should return an error if this minter cannot mint
	// the specified coin.
	OnMint(ctx sdk.Context, minter sdk.AccAddress, coin sdk.Coin) error

	// OnSend specifies a rule for sending coins and should return an error if this sender cannot send
	// the specified coin to the receiver.
	OnSend(ctx sdk.Context, sender sdk.AccAddress, receiver sdk.AccAddress, coin sdk.Coin) error

	// OnBurn specifies a rule for burner coins and should return an error if this burner cannot burn
	// the specified coin.
	OnBurn(ctx sdk.Context, burner sdk.AccAddress, coin sdk.Coin) error
}
