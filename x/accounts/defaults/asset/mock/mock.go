package asset

import (
	"context"

	"cosmossdk.io/math"
	v1 "cosmossdk.io/x/accounts/defaults/asset/v1"
)

var fee = math.LegacyNewDecWithPrec(10, 2) // 10%

func CustomTransfer(aa v1.AssetAccountI) func(ctx context.Context, from, to []byte, amount math.Int) ([][]byte, error) {
	return func(ctx context.Context, from, to []byte, amount math.Int) ([][]byte, error) {
		err := aa.SubUnlockedCoins(ctx, from, amount)
		if err != nil {
			return nil, err
		}

		feeAmount := math.LegacyNewDecFromInt(amount).Mul(fee).TruncateInt()
		transferAmount := amount.Sub(feeAmount)
		owner, err := aa.GetOwner(ctx)
		if err != nil {
			return nil, err
		}

		err = aa.AddCoins(ctx, to, transferAmount)
		if err != nil {
			return nil, err
		}

		err = aa.AddCoins(ctx, owner, feeAmount)
		if err != nil {
			return nil, err
		}

		return [][]byte{from, to, owner}, nil
	}
}

func CustomMint(aa v1.AssetAccountI) func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error) {
	return func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error) {
		feeAmount := math.LegacyNewDecFromInt(amount).Mul(fee).TruncateInt()
		mintAmount := amount.Sub(feeAmount)
		owner, err := aa.GetOwner(ctx)
		if err != nil {
			return nil, err
		}

		err = aa.AddCoins(ctx, to, mintAmount)
		if err != nil {
			return nil, err
		}

		err = aa.AddCoins(ctx, owner, feeAmount)
		if err != nil {
			return nil, err
		}

		return [][]byte{to, owner}, nil
	}
}

func CustomBurn(aa v1.AssetAccountI) func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error) {
	return func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error) {
		feeAmount := math.LegacyNewDecFromInt(amount).Mul(fee).TruncateInt()
		burnAmount := amount.Sub(feeAmount)
		err := aa.SubUnlockedCoins(ctx, from, burnAmount)
		if err != nil {
			return nil, err
		}

		return [][]byte{from}, nil
	}
}
