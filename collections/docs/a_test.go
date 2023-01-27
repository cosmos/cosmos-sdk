package docs

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var AccountsNumberIndexPrefix = collections.NewPrefix(1)

type AccountsIndexes struct {
	Number *indexes.Unique[uint64, sdk.AccAddress, authtypes.BaseAccount]
}

func (a AccountsIndexes) IndexesList() []collections.Index[sdk.AccAddress, authtypes.BaseAccount] {
	return []collections.Index[sdk.AccAddress, authtypes.BaseAccount]{a.Number}
}

func NewAccountIndexes(sb *collections.SchemaBuilder) AccountsIndexes {
	return AccountsIndexes{
		Number: indexes.NewUnique(
			sb, AccountsNumberIndexPrefix, "accounts_by_number",
			collections.Uint64Key, sdk.AccAddressKey,
			func(_ sdk.AccAddress, v authtypes.BaseAccount) (uint64, error) {
				return v.AccountNumber, nil
			},
		),
	}
}

var AccountsPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema   collections.Schema
	Accounts *collections.IndexedMap[sdk.AccAddress, authtypes.BaseAccount, AccountsIndexes]
}

func NewKeeper(storeKey *storetypes.KVStoreKey, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Accounts: collections.NewIndexedMap(
			sb, AccountsPrefix, "accounts",
			sdk.AccAddressKey, codec.CollValue[authtypes.BaseAccount](cdc),
			NewAccountIndexes(sb),
		),
	}
}

func (k Keeper) CreateAccount(ctx sdk.Context, addr sdk.AccAddress) error {
	nextAccountNumber := k.getNextAccountNumber()

	newAcc := authtypes.BaseAccount{
		AccountNumber: nextAccountNumber,
		Sequence:      0,
	}

	return k.Accounts.Set(ctx, addr, newAcc)
}

func (k Keeper) RemoveAccount(ctx sdk.Context, addr sdk.AccAddress) error {
	return k.Accounts.Remove(ctx, addr)
}

func (k Keeper) GetAccountByNumber(ctx sdk.Context, accNumber uint64) (sdk.AccAddress, authtypes.BaseAccount, error) {
	accAddress, err := k.Accounts.Indexes.Number.MatchExact(ctx, accNumber)
	if err != nil {
		return nil, authtypes.BaseAccount{}, err
	}

	acc, err := k.Accounts.Get(ctx, accAddress)
	return accAddress, acc, nil
}

func (k Keeper) GetAccountsByNumber(ctx sdk.Context, startAccNum, endAccNum uint64) ([]authtypes.BaseAccount, error) {
	rng := new(collections.Range[uint64]).
		StartInclusive(startAccNum).
		EndInclusive(endAccNum)

	iter, err := k.Accounts.Indexes.Number.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}

	return indexes.CollectValues(ctx, k.Accounts, iter)
}

func (k Keeper) getNextAccountNumber() uint64 {
	return 0
}
