package valuerenderer

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	corecoins "cosmossdk.io/core/coins"
	"cosmossdk.io/math"
)

const emptyCoins = "zero"

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

	// Since this value renderer has a FormatRepeated method, the Format one
	// here only handles single coin.
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

func (vr coinsValueRenderer) FormatRepeated(ctx context.Context, v protoreflect.Value) ([]Screen, error) {
	if vr.coinMetadataQuerier == nil {
		return nil, fmt.Errorf("expected non-nil coin metadata querier")
	}

	protoCoins := v.List()
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

func (vr coinsValueRenderer) Parse(ctx context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) != 1 {
		return nilValue, fmt.Errorf("expected single screen: %v", screens)
	}

	if screens[0].Text == emptyCoins {
		return protoreflect.ValueOfMessage((&basev1beta1.Coin{}).ProtoReflect()), nil
	}

	parsed, err := vr.parseCoins(ctx, screens[0].Text)
	if err != nil {
		return nilValue, err
	}

	return protoreflect.ValueOfMessage(parsed[0].ProtoReflect()), err
}

func (vr coinsValueRenderer) ParseRepeated(ctx context.Context, screens []Screen, l protoreflect.List) error {
	if len(screens) != 1 {
		return fmt.Errorf("expected single screen: %v", screens)
	}

	if screens[0].Text == emptyCoins {
		return nil
	}

	parsed, err := vr.parseCoins(ctx, screens[0].Text)
	if err != nil {
		return err
	}

	for _, c := range parsed {
		l.Append(protoreflect.ValueOf(c.ProtoReflect()))
	}

	return nil
}

func (vr coinsValueRenderer) parseCoins(ctx context.Context, coinsStr string) ([]*basev1beta1.Coin, error) {
	coins := strings.Split(coinsStr, ", ")
	metadatas := make([]*bankv1beta1.Metadata, len(coins))

	var err error
	for i, coin := range coins {
		coinArr := strings.Split(coin, " ")
		if len(coinArr) != 2 {
			return nil, fmt.Errorf("invalid coin %s", coin)
		}
		metadatas[i], err = vr.coinMetadataQuerier(ctx, coinArr[1])
		if err != nil {
			return nil, err
		}
	}

	if len(coins) != len(metadatas) {
		return []*basev1beta1.Coin{}, fmt.Errorf("formatCoins expect one metadata for each coin; expected %d, got %d", len(coins), len(metadatas))
	}

	parsedCoins := make([]*basev1beta1.Coin, len(coins))
	for i, coinStr := range coins {
		coin, err := parseCoin(coinStr, metadatas[i])
		if err != nil {
			return nil, err
		}
		parsedCoins[i] = coin
	}

	return parsedCoins, nil
}

// parseCoin parses a single value-rendered coin into the Coin struct.
// It shares a lot of code with `cosmos-sdk.io/core/coins.Format`,
// so this code might be refactored once we have
// a core Parse function for coins.
func parseCoin(coinStr string, metadata *bankv1beta1.Metadata) (*basev1beta1.Coin, error) {
	coinArr := strings.Split(coinStr, " ")
	amt1 := coinArr[0]
	coinDenom := coinArr[1]

	if metadata == nil || metadata.Base == "" || coinArr[1] == metadata.Base {
		dec, err := parseDec(amt1)
		if err != nil {
			return nil, err
		}

		return &basev1beta1.Coin{
			Amount: dec,
			Denom:  coinDenom,
		}, nil
	}
	baseDenom := metadata.Base

	// Find exponents of both denoms.
	foundCoinExp, foundBaseExp := false, false
	var coinExp, baseExp uint32
	for _, unit := range metadata.DenomUnits {
		if coinDenom == unit.Denom {
			coinExp = unit.Exponent
			foundCoinExp = true
		}
		if baseDenom == unit.Denom {
			baseExp = unit.Exponent
			foundBaseExp = true
		}
	}

	// If we didn't find either exponent, then we return early.
	if !foundCoinExp || !foundBaseExp {
		amt, err := parseDec(amt1)
		if err != nil {
			return nil, err
		}

		return &basev1beta1.Coin{
			Amount: amt,
			Denom:  baseDenom,
		}, nil
	}

	// remove 1000 separators, (ex: 1'000'000 -> 1000000)
	amt1 = strings.ReplaceAll(amt1, "'", "")
	amt, err := math.LegacyNewDecFromStr(amt1)
	if err != nil {
		return nil, err
	}

	if coinExp > baseExp {
		amt = amt.Mul(math.LegacyNewDec(10).Power(uint64(coinExp - baseExp)))
	} else {
		amt = amt.Quo(math.LegacyNewDec(10).Power(uint64(baseExp - coinExp)))
	}

	amtStr, err := parseDec(amt.String())
	if err != nil {
		return nil, err
	}

	return &basev1beta1.Coin{
		Amount: amtStr,
		Denom:  baseDenom,
	}, nil
}
