package keeper

import (
	"bytes"
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accountlink/interfaces"
	"cosmossdk.io/x/accountlink/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	gogoproto "github.com/cosmos/gogoproto/proto"
)

// ModuleAccountAddress defines the x/accounts module address.
var ModuleAccountAddress = address.Module(types.ModuleName)

type LegacyStateCodec interface {
	Marshal(gogoproto.Message) ([]byte, error)
	Unmarshal([]byte, gogoproto.Message) error
}

type AccountsIndexes struct {
	AccountType *indexes.ReversePair[sdk.AccAddress, string, types.AccountsMetadata]
}

func newAccountsIndexes(sb *collections.SchemaBuilder) AccountsIndexes {
	return AccountsIndexes{
		AccountType: indexes.NewReversePair[types.AccountsMetadata](
			sb, types.AccountTypeAddressPrefix, "addresses_by_account_type_index",
			collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey),
			indexes.WithReversePairUncheckedValue(),
		),
	}
}

func (a AccountsIndexes) IndexesList() []collections.Index[collections.Pair[sdk.AccAddress, string], types.AccountsMetadata] {
	return []collections.Index[collections.Pair[sdk.AccAddress, string], types.AccountsMetadata]{a.AccountType}
}

type Keeper struct {
	authKeeper    types.AuthKeeper
	accountKeeper types.AccountKeeper
	// map all smart account addresses by key pair of owner address and account type
	accounts collections.IndexedMap[collections.Pair[sdk.AccAddress, string], types.AccountsMetadata, AccountsIndexes]
}

func NewKeeper(authKeeper types.AuthKeeper, accountKeeper types.AccountKeeper, storeService store.KVStoreService, cdc LegacyStateCodec) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return Keeper{
		authKeeper: authKeeper,
		accounts:   *collections.NewIndexedMap(sb, types.AccountsPrefix, "accounts", collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), codec.CollValue[types.AccountsMetadata](cdc), newAccountsIndexes(sb)),
	}
}

func (k Keeper) GetAccountsByOwner(ctx context.Context, owner sdk.AccAddress, accountType string) types.AccountsMetadata {
	addresses, err := k.accounts.Get(ctx, collections.Join(owner, accountType))
	if err != nil {
		return types.AccountsMetadata{
			Addresses: []string{},
		}
	}
	return addresses
}

func (k Keeper) SetAccountsByOwner(ctx context.Context, owner sdk.AccAddress, accountType, address string) error {
	accounts := k.GetAccountsByOwner(ctx, owner, accountType)
	accounts.Addresses = append(accounts.Addresses, address)

	err := k.accounts.Set(ctx, collections.Join(owner, accountType), accounts)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) CheckCondition(ctx context.Context, condition types.Condition) error {
	byteAccAddr, err := k.authKeeper.AddressCodec().StringToBytes(condition.Account)
	if err != nil {
		return err
	}

	_, err = k.accountKeeper.Execute(ctx, byteAccAddr, ModuleAccountAddress, &interfaces.MsgConditionCheck{
		Sender:    condition.Owner,
		Condition: &condition,
	})

	// here is where we check if the account handles condition check messages
	// if it does not, then we simply skip the condition check
	switch {
	case err == nil, types.IsRoutingError(err):
		// if we get a routing message error it means the account does not handle condition check messages,
		// in this case we then we simply skip the condition check
		return nil
	default:
		// some other execution error.
		return err
	}
}

func (k Keeper) IsAccountModule(addr []byte) bool {
	return bytes.Equal(addr, ModuleAccountAddress)
}
