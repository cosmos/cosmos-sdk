package valuerenderer

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type coinsValueRenderer struct {
	// coinMetadataQuerier defines a function to query the coin metadata from
	// state.
	coinMetadataQuerier CoinMetadataQueryFn
}

var _ ValueRenderer = coinsValueRenderer{}

func (vr coinsValueRenderer) Format(ctx context.Context, v protoreflect.Value, w io.Writer) error {
	if vr.coinMetadataQuerier == nil {
		return fmt.Errorf("expected non-nil coin metadata querier")
	}

	coins := v.Interface().([]*basev1beta1.Coin)

	metadatas := make([]*bankv1beta1.Metadata, len(coins))
	var err error
	for i, coin := range coins {
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

func (vr coinsValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	panic("implement me")
}

// formatCoins formats Coins into a value-rendered string, which uses
// `formatCoin` separated by ", " (a comma and a space), and sorted
// alphabetically by value-rendered denoms. It expects an array of metadata
// (optionally nil), where each metadata at index `i` MUST match the coin denom
// at the same index.
func formatCoins(coins []*basev1beta1.Coin, metadata []*bankv1beta1.Metadata) (string, error) {
	if len(coins) != len(metadata) {
		return "", fmt.Errorf("formatCoins expect one metadata for each coin; expected %d, got %d", len(coins), len(metadata))
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
