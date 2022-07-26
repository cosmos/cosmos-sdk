package valuerenderer

import (
	"context"
	"fmt"
	"io"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/math"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type coinValueRenderer struct {
	// coinMetadataQuerier defines a function to query the coin metadata from
	// state.
	coinMetadataQuerier CoinMetadataQueryFn
}

var _ ValueRenderer = coinValueRenderer{}

func (vr coinValueRenderer) Format(ctx context.Context, v protoreflect.Value, w io.Writer) error {
	if vr.coinMetadataQuerier == nil {
		return fmt.Errorf("expected non-nil coin metadata querier")
	}

	coin := v.Interface().(protoreflect.Message).Interface().(*basev1beta1.Coin)

	metadata, err := vr.coinMetadataQuerier(ctx, coin.Denom)
	if err != nil {
		return err
	}

	formatted, err := formatCoin(coin, metadata)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(formatted))
	return err
}

func (vr coinValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	panic("implement me")
}

// formatCoin formats a sdk.Coin into a value-rendered string, using the
// given metadata about the denom. It returns the formatted coin string, the
// display denom, and an optional error.
func formatCoin(coin *basev1beta1.Coin, metadata *bankv1beta1.Metadata) (string, error) {
	coinDenom := coin.Denom

	// Return early if no display denom or display denom is the current coin denom.
	if metadata == nil || metadata.Display == "" || coinDenom == metadata.Display {
		vr, err := formatDecimal(coin.Amount)
		return vr + " " + coin.Denom, err
	}

	dispDenom := metadata.Display

	// Find exponents of both denoms.
	var coinExp, dispExp uint32
	foundCoinExp, foundDispExp := false, false
	for _, unit := range metadata.DenomUnits {
		if coinDenom == unit.Denom {
			coinExp = unit.Exponent
			foundCoinExp = true
		}
		if dispDenom == unit.Denom {
			dispExp = unit.Exponent
			foundDispExp = true
		}
	}

	// If we didn't find either exponent, then we return early.
	if !foundCoinExp || !foundDispExp {
		vr, err := formatInteger(coin.Amount)
		return vr + " " + coin.Denom, err
	}

	exponentDiff := int64(coinExp) - int64(dispExp)
	var (
		dispAmount math.LegacyDec
		err        error
	)
	if exponentDiff > 0 {
		dispAmount, err = math.LegacyNewDecFromStr(coin.Amount)
		if err != nil {
			return "", err
		}
		dispAmount = dispAmount.Mul(math.LegacyNewDec(10).Power(uint64(exponentDiff)))
	} else {
		dispAmount, err = math.LegacyNewDecFromStr(coin.Amount)
		if err != nil {
			return "", err
		}
		dispAmount = dispAmount.Quo(math.LegacyNewDec(10).Power(uint64(-exponentDiff)))
	}

	vr, err := formatDecimal(dispAmount.String())
	return vr + " " + dispDenom, err
}
