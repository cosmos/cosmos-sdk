package common

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// AccountSource provides a method to retrieve an account based on its address.
	AccountSource interface {
		GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	}

	// ModuleAccountSource defines a method to retrieve the address of a module account by its module name.
	ModuleAccountSource interface {
		GetModuleAddress(moduleName string) sdk.AccAddress
	}

	// AccountSourceX Account and Module account
	AccountSourceX interface {
		AccountSource
		ModuleAccountSource
	}
	// SimDeliveryResultHandler processes the delivery response error. Some sims are supposed to fail and expect an error.
	// An unhandled error returned indicates a failure
	SimDeliveryResultHandler func(error) error
)
