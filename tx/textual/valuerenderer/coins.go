package valuerenderer

import (
	"context"
	"fmt"

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

func (vr coinsValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	// ref: https://github.com/cosmos/cosmos-sdk/issues/13153
	panic("implement me, see #13153")
}
