package valuerenderer

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	corecoins "cosmossdk.io/core/coins"
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

func (vr coinsValueRenderer) Format(ctx context.Context, v protoreflect.Value) ([]Screen, error) {
	if vr.coinMetadataQuerier == nil {
		return nil, fmt.Errorf("expected non-nil coin metadata querier")
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
					return nil, err
				}
			}

			formatted, err := corecoins.FormatCoins(coins, metadatas)
			if err != nil {
				return nil, err
			}

			return []Screen{{Text: formatted}}, nil
		}
	// If it's a single Coin:
	case protoreflect.Message:
		{
			coin := v.Interface().(protoreflect.Message).Interface().(*basev1beta1.Coin)

			metadata, err := vr.coinMetadataQuerier(ctx, coin.Denom)
			if err != nil {
				return nil, err
			}

			formatted, err := corecoins.FormatCoins([]*basev1beta1.Coin{coin}, []*bankv1beta1.Metadata{metadata})
			if err != nil {
				return nil, err
			}

			return []Screen{{Text: formatted}}, nil
		}
	default:
		return nil, fmt.Errorf("got invalid type %t for coins", v.Interface())
	}
}

func (vr coinsValueRenderer) Parse(ctx context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) != 1 {
		return protoreflect.Value{}, fmt.Errorf("expected single screen: %v", screens)
	}

	if screens[0].Text == corecoins.EmptyCoins {
		return protoreflect.ValueOfList(NewGenericList([]*basev1beta1.Coin{})), nil
	}

	coins := strings.Split(screens[0].Text, ", ")
	metadatas := make([]*bankv1beta1.Metadata, len(coins))

	var err error
	for i, coin := range coins {
		coinArr := strings.Split(coin, " ")
		if len(coinArr) != 2 {
			return protoreflect.Value{}, fmt.Errorf("invalid coin %s", coins)
		}
		metadatas[i], err = vr.coinMetadataQuerier(ctx, coinArr[1])
		if err != nil {
			return protoreflect.Value{}, err
		}
	}

	parsed, err := parseCoins(coins, metadatas)
	if err != nil {
		return protoreflect.Value{}, err
	}

	if len(parsed) > 1 {
		return protoreflect.ValueOf(NewGenericList(parsed)), err
	} else {
		return protoreflect.ValueOfMessage(parsed[0].ProtoReflect()), err
	}
}

func parseCoins(coins []string, metadata []*bankv1beta1.Metadata) ([]*basev1beta1.Coin, error) {
	if len(coins) != len(metadata) {
		return []*basev1beta1.Coin{}, fmt.Errorf("formatCoins expect one metadata for each coin; expected %d, got %d", len(coins), len(metadata))
	}

	parsedCoins := make([]*basev1beta1.Coin, len(coins))
	for i, coinStr := range coins {
		coin, err := parseCoin(coinStr, metadata[i])
		if err != nil {
			return []*basev1beta1.Coin{}, err
		}
		parsedCoins[i] = coin
	}

	return parsedCoins, nil
}

func parseCoin(coinStr string, metadata *bankv1beta1.Metadata) (*basev1beta1.Coin, error) {
	coinArr := strings.Split(coinStr, " ")
	amt1 := strings.Replace(coinArr[0], "'", "", -1)
	coinDenom := coinArr[1]

	var base = coinDenom
	if metadata != nil {
		base = metadata.Base
	}

	vr, denom, err := corecoins.ConvertAmt(
		&basev1beta1.Coin{
			Amount: amt1,
			Denom:  coinDenom,
		}, metadata, base,
	)

	// vr value contains 1000 separators, need to be removed.
	vr = strings.Replace(vr, "'", "", -1)

	return &basev1beta1.Coin{
		Denom:  denom,
		Amount: vr,
	}, err
}
