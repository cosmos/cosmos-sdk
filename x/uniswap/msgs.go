package uniswap

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const RouterKey = ModuleName

// MsgCreateExchange - add a new trading pair
type MsgCreateExchange struct {
	NewCoin string
}

var _ sdk.Msg = MsgCreateExchange{}

// NewCreateExchange - .
func NewMsgCreateExchange(newCoin string) MsgCreateExchange {
	return MsgSend{NewCoin: newCoin}
}

// Route Implements Msg.
func (msg MsgCreateExchange) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgCreateExchange) Type() string { return "create_exchange" }

// ValidateBasic Implements Msg.
func (msg MsgCreateExchange) ValidateBasic() sdk.Error {
	if !(len(msg.NewCoin) > 0) {
		return errors.New("must provide coin denomination")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgCreateExchange) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgCreateExchange) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

type MsgSwapOrder struct {
	SwapDenom  string         // The desired denomination either to be bought or sold
	Coin       sdk.Coin       // The specified amount to be either bought or sold
	Bound      sdk.Int        // If buy order, maximum amount of coins to be sold; otherwise minimum amount of coins to be bought
	Deadline   time.Time      // deadline for the transaction to still be considered valid
	Recipient  sdk.AccAddress // address output coin is being sent to
	IsBuyOrder bool           // boolean indicating whether the order should be treated as a buy or sell
}

type MsgAddLiquidity struct {
	DepositAmount sdk.Int // exact amount of native asset being add to the liquidity pool
	ExchangeDenom string  // denomination of the exchange being added to
	MinLiquidity  sdk.Int // lower bound UNI sender is willing to accept for deposited coins
	MaxCoins      sdk.Int // maximum amount of the coin the sender is willing to deposit.
	Deadline      time.Time
}

type MsgRemoveLiquidity struct {
	WithdrawAmount sdk.Int // amount of UNI to be burned to withdraw liquidity from an exchange
	ExchangeDenom  string  // denomination of the exchange being withdrawn from
	MinNative      sdk.Int // minimum amount of the native asset the sender is willing to accept
	MinCoins       sdk.Int // minimum amount of the exchange coin the sender is willing to accept
	Deadline       time.Time
}
