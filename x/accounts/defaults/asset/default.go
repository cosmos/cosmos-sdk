package asset

import (
	"context"

	"cosmossdk.io/math"
	v1 "cosmossdk.io/x/accounts/defaults/asset/v1"
)

// DefaultTransfer same with sdk Transfer logic
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

// DefaultMint same with sdk Mint logic
func DefaultMint(aa v1.AssetAccountI) func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error) {
	return func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error) {
		err := aa.AddCoins(ctx, to, amount)
		if err != nil {
			return nil, err
		}

		supply := aa.GetSupply(ctx)
		supply = supply.Add(amount)
		err = aa.SetSupply(ctx, supply)
		if err != nil {
			return nil, err
		}

		return [][]byte{to}, nil
	}
}

// DefaultBurn same with sdk Burn logic
func DefaultBurn(aa v1.AssetAccountI) func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error) {
	return func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error) {
		err := aa.SubUnlockedCoins(ctx, from, amount)
		if err != nil {
			return nil, err
		}

		supply := aa.GetSupply(ctx)
		supply, err = supply.SafeSub(amount)
		if err != nil {
			return nil, err
		}

		err = aa.SetSupply(ctx, supply)
		if err != nil {
			return nil, err
		}

		return [][]byte{from}, nil
	}
}
