package uniswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

var (
	_ sdk.Msg = MsgSwapOrder{}
	_ sdk.Msg = MsgAddLiquidity{}
	_ sdk.Msg = MsgRemoveLiquidity{}
)

/* --------------------------------------------------------------------------- */
// MsgSwapOrder
/* --------------------------------------------------------------------------- */

// MsgSwap Order - struct for swapping a coin
type MsgSwapOrder struct {
	SwapDenom  string         // The desired denomination either to be bought or sold
	Coin       sdk.Coin       // The specified amount to be either bought or sold
	Bound      sdk.Int        // If buy order, maximum amount of coins to be sold; otherwise minimum amount of coins to be bought
	Deadline   time.Time      // deadline for the transaction to still be considered valid
	Sender     sdk.AccAddress // address swapping coin
	Recipient  sdk.AccAddress // address output coin is being sent to
	IsBuyOrder bool           // boolean indicating whether the order should be treated as a buy or sell
}

// NewMsgSwapOrder is a constructor function for MsgSwapOrder
func NewMsgSwapOrder(
	swapDenom string, coin sdk.Coin, bound sdk.Int, deadline time.Time,
	sender, recipient sdk.AccAddress, isBuyOrder bool,
) MsgSwapOrder {

	return MsgSwapOrder{
		SwapDenom:  swapDenom,
		Coin:       coin,
		Bound:      bound,
		Deadline:   deadline,
		Sender:     sender,
		Recipient:  recipient,
		IsBuyOrder: isBuyOrder,
	}
}

// Type Implements Msg
func (msg MsgSwapOrder) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgSwapOrder) Type() string { return "swap_order" }

// ValidateBasic Implements Msg.
func (msg MsgSwapOrder) ValidateBasic() sdk.Error {
	if msg.SwapDenom == "" {
		return ErrNoDenom(DefaultCodespace)
	}
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress("invalid sender address")
	}
	if msg.Recipient.Empty() {
		return sdk.ErrInvalidAddress("invalid recipient address")
	}
	if msg.Coin.IsZero() {
		return ErrZeroCoin()
	}
}

// GetSignBytes Implements Msg.
func (msg MsgSwapOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgSwapOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

/* --------------------------------------------------------------------------- */
// MsgAddLiquidity
/* --------------------------------------------------------------------------- */

// MsgAddLiquidity - struct for adding liquidity to an exchange
type MsgAddLiquidity struct {
	DepositAmount sdk.Int // exact amount of native asset being add to the liquidity pool
	ExchangeDenom string  // denomination of the exchange being added to
	MinLiquidity  sdk.Int // lower bound UNI sender is willing to accept for deposited coins
	MaxCoins      sdk.Int // maximum amount of the coin the sender is willing to deposit.
	Deadline      time.Time
	Sender        sdk.AccAddress
}

// NewMsgAddLiquidity is a constructor function for MsgAddLiquidity
func NewMsgAddLiquidity(
	depositAmount, minLiquidity, maxCoins sdk.Int,
	exchangeDenom string, deadline time.Time, sender sdk.AccAddress,
) MsgAddLiquidity {

	return MsgAddLiquidity{
		DepositAmount: depositAmount,
		ExchangeDenom: exchangeDenom,
		MinLiquidity:  minLiquidity,
		MaxCoins:      maxCoins,
		Deadline:      deadline,
		Sender:        sender,
	}
}

// Type Implements Msg
func (msg MsgAddLiquidity) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgAddLiquidity) Type() string { return "add_liquidity" }

// ValidateBasic Implements Msg.
func (msg MsgAddLiquidity) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.InvalidAddress("invalid sender address")
	}
}

// GetSignBytes Implements Msg.
func (msg MsgAddLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

/* --------------------------------------------------------------------------- */
// MsgRemoveLiquidity
/* --------------------------------------------------------------------------- */

// MsgRemoveLiquidity - struct for removing liquidity from an exchange
type MsgRemoveLiquidity struct {
	WithdrawAmount sdk.Int // amount of UNI to be burned to withdraw liquidity from an exchange
	ExchangeDenom  string  // denomination of the exchange being withdrawn from
	MinNative      sdk.Int // minimum amount of the native asset the sender is willing to accept
	MinCoins       sdk.Int // minimum amount of the exchange coin the sender is willing to accept
	Deadline       time.Time
	Sender         sdk.AccAddress
}

// NewMsgRemoveLiquidity is a contructor function for MsgRemoveLiquidity
func NewMsgRemoveLiquidity(
	withdrawAmount, minNative, minCoins sdk.Int,
	exchangeDenom string, deadline time.Time, sender sdk.AccAddress,
) MsgRemoveLiquidity {

	return MsgRemoveLiquidity{
		WithdrawAmount: withdrawAmount,
		ExchangeDenom:  exchangeDenom,
		MinNative:      minNative,
		MinCoins:       minCoins,
		Deadline:       deadline,
		Sender:         sender,
	}
}

// Type Implements Msg
func (msg MsgRemoveLiquidity) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgRemoveLiquidity) Type() string { return "remove_liquidity" }

// ValidateBasic Implements Msg.
func (msg MsgRemoveLiquidity) ValidateBasic() sdk.Error {
	if msg.ExchangeDenom == "" {
		return ErrNoDenom(DefaultCodespace)
	}
	if msg.Sender.Empty() {
		return sdk.InvalidAddress("invalid sender address")
	}
}

// GetSignBytes Implements Msg.
func (msg MsgRemoveLiquidity) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgRemoveLiquidity) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
