package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type legacyBalance struct {
	addr  sdk.AccAddress
	denom string
	amt   math.Int
}

// MigrateLegacyBalances rewrites balances still stored as Coin protobuf
// and creates the missing denom to address index rows.
// Balances already stored as an Int are left untouched.
// A second run rewrites nothing.
func (k BaseKeeper) MigrateLegacyBalances(ctx context.Context) (int, error) {
	kv := k.storeService.OpenKVStore(ctx)
	start := types.BalancesPrefix.Bytes()
	end := make([]byte, len(start))
	copy(end, start)
	end[len(end)-1]++

	iter, err := kv.Iterator(start, end)
	if err != nil {
		return 0, err
	}
	closed := false
	closeIter := func() error {
		if closed {
			return nil
		}
		closed = true
		return iter.Close()
	}
	defer func() {
		_ = closeIter()
	}()

	var pending []legacyBalance
	for ; iter.Valid(); iter.Next() {
		value := append([]byte(nil), iter.Value()...)
		if _, err := sdk.IntValue.Decode(value); err == nil {
			continue
		}

		coin := new(sdk.Coin)
		if err := coin.Unmarshal(value); err != nil {
			return 0, fmt.Errorf("balance value is neither an int nor a coin: %w", err)
		}
		if coin.Denom == "" || coin.Amount.IsNil() {
			return 0, fmt.Errorf("balance value is neither an int nor a coin")
		}

		rawKey := append([]byte(nil), iter.Key()...)
		if len(rawKey) < len(start) {
			return 0, fmt.Errorf("balance key shorter than prefix")
		}
		n, pk, err := k.Balances.KeyCodec().Decode(rawKey[len(start):])
		if err != nil {
			return 0, fmt.Errorf("decode balance key: %w", err)
		}
		if n != len(rawKey[len(start):]) {
			return 0, fmt.Errorf("decode balance key: trailing bytes")
		}
		if coin.Denom != pk.K2() {
			return 0, fmt.Errorf("legacy balance denom %s does not match key denom %s", coin.Denom, pk.K2())
		}

		pending = append(pending, legacyBalance{
			addr:  sdk.AccAddress(append([]byte(nil), pk.K1()...)),
			denom: pk.K2(),
			amt:   coin.Amount,
		})
	}
	if err := closeIter(); err != nil {
		return 0, err
	}

	for _, bal := range pending {
		if err := k.Balances.Set(ctx, collections.Join(bal.addr, bal.denom), bal.amt); err != nil {
			return 0, err
		}
	}
	return len(pending), nil
}
