package types

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DenomResolver is an interface to convert a given token to the feemarket's base token.
type DenomResolver interface {
	// ConvertToDenom converts deccoin into the equivalent amount of the token denominated in denom.
	ConvertToDenom(ctx context.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error)
	// ExtraDenoms returns a list of denoms in addition of `Params.base_denom` it's possible to pay fees with
	ExtraDenoms(ctx context.Context) ([]string, error)
}

// TestDenomResolver is a test implementation of the DenomResolver interface.  It returns "feeCoin.Amount baseDenom" for all coins that are not the baseDenom.
// NOTE: DO NOT USE THIS IN PRODUCTION
type TestDenomResolver struct{}

// ConvertToDenom returns "coin.Amount denom" for all coins that are not the denom.
func (r *TestDenomResolver) ConvertToDenom(_ context.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error) {
	if coin.Denom == denom {
		return coin, nil
	}

	return sdk.NewDecCoinFromDec(denom, coin.Amount), nil
}

func (r *TestDenomResolver) ExtraDenoms(_ context.Context) ([]string, error) {
	return []string{}, nil
}

// ErrorDenomResolver is a test implementation of the DenomResolver interface.  It returns an error for all coins that are not the baseDenom.
// NOTE: DO NOT USE THIS IN PRODUCTION
type ErrorDenomResolver struct{}

// ConvertToDenom returns an error for all coins that are not the denom.
func (r *ErrorDenomResolver) ConvertToDenom(_ context.Context, coin sdk.DecCoin, denom string) (sdk.DecCoin, error) {
	if coin.Denom == denom {
		return coin, nil
	}

	return sdk.DecCoin{}, fmt.Errorf("error resolving denom")
}

func (r *ErrorDenomResolver) ExtraDenoms(_ context.Context) ([]string, error) {
	return []string{}, nil
}
