package auth

import (
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

type (
	Account = exported.Account
	ObserverI = keeper.ObserverI
)
