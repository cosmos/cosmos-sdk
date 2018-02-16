package coins

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
)

// CoinMapper manages transfers between accounts
type coinMapper struct {
	// Reference to underlying AccountMapper
	am *sdk.AccountMapper

	// The (unexposed) key used to access the store from the Context.
	key sdk.StoreKey

	// The wire codec for binary encoding/decoding of accounts.
	cdc *wire.Codec
}

// NewCoinMapper returns a new CoinMapper that
// uses go-wire to (binary) encode and decode CoinAccounts.
func NewCoinMapper(key sdk.StoreKey, am auth.AccountMapper) coinMapper {
	cdc := wire.NewCodec()
	return coinMapper{
		am:  am,
		key: key,
		cdc: cdc,
	}
}

// SubtractCoins subtracts amt from the coins at the addr.
func (cm CoinMapper) MakeCoinAccount(ctx sdk.Context, acc auth.BaseAccount, amt Coins) (CoinAccount, sdk.Error) {
	addr := acc.GetAddress(ctx, acc)

	coinAcc := NewCoinAccountWithAccount(acc)
	coinAcc.SetCoins(coins)

	cm.SetCoinAccount(ctx, coinAcc)
	return coinAcc, nil
}

// GetCoinAccount reads a CoinAccount from store
func (cm CoinMapper) GetCoinAccount(ctx sdk.Context, acc auth.BaseAccount) CoinAccount {
	addr := acc.GetAddress()

	store := ctx.KVStore(cm.key)
	bz := store.Get(addr)
	if bz == nil {
		return nil
	}
	coinAcc := cm.decodeCoinAccount(bz)

	return coinAcc
}

// SetCoinAccount writes a CoinAccount to store
func (cm CoinMapper) SetCoinAccount(ctx sdk.Context, coinAcc CoinAccount) {
	addr := coinAcc.GetAddress()

	if cm.GetAccount(addr) != coinAcc.Account {
		cm.SetAccount(ctx, coinAcc.Account)
	}

	coinStore := ctx.KVStore(cm.key)
	bz := cm.encodeCoinAccount(acc)
	store.Set(addr, bz)
}
