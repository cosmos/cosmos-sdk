package token

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

type Keeper struct {
	bank    bank.Keeper
	channel channel.Manager
}

func NewKeeper(bank bank.Keeper, channel channel.Manager) Keeper {
	return Keeper{
		bank:    bank,
		channel: channel,
	}
}

func (k Keeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, toConn, toChan string, amt sdk.Coins) sdk.Error {
	_, err := k.bank.SubtractCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}
	err = k.channel.Send(ctx, toConn, toChan, PacketSend{toAddr, amt})
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) receiveCoins(ctx sdk.Context, toAddr sdk.AccAddress, amt sdk.Coins) sdk.Error {
	_, err := k.bank.AddCoins(ctx, toAddr, amt)
	if err != nil {
		return err
		// TODO: should return receipt
	}

	return nil
}
