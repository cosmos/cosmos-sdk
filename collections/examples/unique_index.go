package examples

import (
	"context"
	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
)

var accountCodec collections.ValueCodec[Account] = nil

type Account struct {
	Number uint64
	PubKey []byte
}

type AccountIndexes struct {
	// Number is the index that maps accounts by number.
	// Since account number is unique for every account we use an indexes.Unique
	Number *indexes.Unique[uint64, AccAddress, Account]
}

func (i AccountIndexes) IndexesList() []collections.Index[AccAddress, Account] {
	return []collections.Index[AccAddress, Account]{i.Number}
}

type Keeper3 struct {
	Accounts *collections.IndexedMap[AccAddress, Account, AccountIndexes]
}

func NewKeeper3() Keeper3 {
	sb := collections.NewSchemaBuilder(nil)
	return Keeper3{
		Accounts: collections.NewIndexedMap(
			sb, collections.NewPrefix(0), "accounts",
			collections.StringKey, // primary key codec, AccAddress is a string alias.
			accountCodec,          // faux value codec.
			AccountIndexes{
				Number: indexes.NewUnique(
					sb, collections.NewPrefix(1), "accounts_by_number",
					collections.Uint64Key, // the account number is uint64, so we use a uint64 key.
					collections.StringKey, // the primary key we use to save accounts is AccAddress which is string.
					func(_ AccAddress, v Account) (uint64, error) { // now we pass the function that given an account returns the acc number
						return v.Number, nil
					},
				),
			},
		),
	}
}

func (k Keeper3) GetAccountByNumber(ctx context.Context, accNum uint64) (Account, error) {
	addr, err := k.Accounts.Indexes.Number.MatchExact(ctx, accNum)
	if err != nil {
		return Account{}, err
	}
	return k.Accounts.Get(ctx, addr)
}
