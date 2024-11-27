package asset

import (
	"context"

	"cosmossdk.io/math"
	v1 "cosmossdk.io/x/accounts/defaults/asset/v1"
)

func DefaultTransfer(aa v1.AssetAccountI) func(ctx context.Context, from, to []byte, amount math.Int) error {
	return func(ctx context.Context, from, to []byte, amount math.Int) error {
		err := aa.SubUnlockedCoins(ctx, from, amount)
		if err != nil {
			return err
		}

		err = aa.AddCoins(ctx, to, amount)
		if err != nil {
			return err
		}

		return nil
	}
}
