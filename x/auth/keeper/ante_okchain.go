package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

type ValidateMsgHandler func(ctx sdk.Context, msgs []sdk.Msg) sdk.Result

type IsSystemFreeHandler func(ctx sdk.Context, msgs []sdk.Msg) bool


type ObserverI interface {
	OnAccountUpdated(acc exported.Account)
}

func (k *AccountKeeper) SetObserverKeeper(observer ObserverI) {
	k.observers = append(k.observers, observer)
}