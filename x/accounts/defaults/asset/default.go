package asset

import (
	"context"

	"cosmossdk.io/math"
	v1 "cosmossdk.io/x/accounts/defaults/asset/v1"
)

func DefaultTransfer(aa v1.AssetAccountI) func(ctx context.Context, from, to []byte, amount math.Int) ([][]byte, error) {
	return func(ctx context.Context, from, to []byte, amount math.Int) ([][]byte, error) {
		err := aa.SubUnlockedCoins(ctx, from, amount)
		if err != nil {
			return nil, err
		}

		err = aa.AddCoins(ctx, to, amount)
		if err != nil {
			return nil, err
		}

		return [][]byte{from, to}, nil
	}
}

func DefaultMint(aa v1.AssetAccountI) func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error) {
	return func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error) {
		err := aa.AddCoins(ctx, to, amount)
		if err != nil {
			return nil, err
		}

		return [][]byte{to}, nil
	}
}

func DefaultBurn(aa v1.AssetAccountI) func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error) {
	return func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error) {
		err := aa.SubUnlockedCoins(ctx, from, amount)
		if err != nil {
			return nil, err
		}

		return [][]byte{from}, nil
	}
}

func CustomTransfer(aa v1.AssetAccountI) func(ctx context.Context, from, to []byte, amount math.Int) error {
	return func(ctx context.Context, from, to []byte, amount math.Int) error {
		err := aa.SubUnlockedCoins(ctx, from, amount)
		if err != nil {
			return err
		}

		fee := math.LegacyNewDecWithPrec(10, 2) // 10%
		feeAmount := math.LegacyNewDecFromInt(amount).Mul(fee).TruncateInt()
		transferAmount := amount.Sub(feeAmount)
		owner, err := aa.GetOwner(ctx)
		if err != nil {
			return err
		}

		err = aa.AddCoins(ctx, to, transferAmount)
		if err != nil {
			return err
		}

		err = aa.AddCoins(ctx, owner, feeAmount)
		if err != nil {
			return err
		}

		return nil
	}
}
