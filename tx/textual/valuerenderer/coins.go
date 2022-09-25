package valuerenderer

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/math"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// NewCoinsValueRenderer returns a ValueRenderer for SDK Coin and Coins.
func NewCoinsValueRenderer(q CoinMetadataQueryFn) ValueRenderer {
	return coinsValueRenderer{q}
}

type coinsValueRenderer struct {
	// coinMetadataQuerier defines a function to query the coin metadata from
	// state. It should use bank module's `DenomsMetadata` gRPC query to fetch
	// each denom's associated metadata, either using the bank keeper (for
	// server-side code) or a gRPC query client (for client-side code).
	coinMetadataQuerier CoinMetadataQueryFn
}

var _ ValueRenderer = coinsValueRenderer{}

func (vr coinsValueRenderer) Format(ctx context.Context, v protoreflect.Value, w io.Writer) error {
	if vr.coinMetadataQuerier == nil {
		return fmt.Errorf("expected non-nil coin metadata querier")
	}

	// Check whether we have a Coin or some Coins.
	switch protoCoins := v.Interface().(type) {
	// If it's a repeated Coin:
	case protoreflect.List:
		{
			coins, metadatas := make([]*basev1beta1.Coin, protoCoins.Len()), make([]*bankv1beta1.Metadata, protoCoins.Len())
			var err error
			for i := 0; i < protoCoins.Len(); i++ {
				coin := protoCoins.Get(i).Interface().(protoreflect.Message).Interface().(*basev1beta1.Coin)
				coins[i] = coin
				metadatas[i], err = vr.coinMetadataQuerier(ctx, coin.Denom)
				if err != nil {
					return err
				}
			}

			formatted, err := formatCoins(coins, metadatas)
			if err != nil {
				return err
			}

			_, err = w.Write([]byte(formatted))
			return err
		}
	// If it's a single Coin:
	case protoreflect.Message:
		{
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
	default:
		return fmt.Errorf("got invalid type %t for coins", v.Interface())
	}
}

func (vr coinsValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	// ref: https://github.com/cosmos/cosmos-sdk/issues/13153
	panic("implement me, see #13153")
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

	dispAmount, err := math.LegacyNewDecFromStr(coin.Amount)
	if err != nil {
		return "", err
	}

	if exponentDiff > 0 {
		dispAmount = dispAmount.Mul(math.LegacyNewDec(10).Power(uint64(exponentDiff)))
	} else {
		dispAmount = dispAmount.Quo(math.LegacyNewDec(10).Power(uint64(-exponentDiff)))
	}

	vr, err := formatDecimal(dispAmount.String())
	return vr + " " + dispDenom, err
}

// formatCoins formats Coins into a value-rendered string, which uses
// `formatCoin` separated by ", " (a comma and a space), and sorted
// alphabetically by value-rendered denoms. It expects an array of metadata
// (optionally nil), where each metadata at index `i` MUST match the coin denom
// at the same index.
func formatCoins(coins []*basev1beta1.Coin, metadata []*bankv1beta1.Metadata) (string, error) {
	if len(coins) != len(metadata) {
		panic(fmt.Errorf("formatCoins expect one metadata for each coin; expected %d, got %d", len(coins), len(metadata)))
	}

	formatted := make([]string, len(coins))
	for i, coin := range coins {
		var err error
		formatted[i], err = formatCoin(coin, metadata[i])
		if err != nil {
			return "", err
		}
	}

	// Sort the formatted coins by display denom.
	sort.SliceStable(formatted, func(i, j int) bool {
		denomI := strings.Split(formatted[i], " ")[1]
		denomJ := strings.Split(formatted[j], " ")[1]

		return denomI < denomJ
	})

	return strings.Join(formatted, ", "), nil
}
