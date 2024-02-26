package keeper

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
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

type Keeper struct {
	authKeeper    types.AuthKeeper
	accountKeeper types.AccountKeeper
	// map all smart account addresses by key pair of owner address and account type
	accounts collections.Map[collections.Pair[sdk.AccAddress, string], types.AccountsMetadata]
}

func NewKeeper(authKeeper types.AuthKeeper, accountKeeper types.AccountKeeper, storeService store.KVStoreService, cdc LegacyStateCodec) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return Keeper{
		authKeeper: authKeeper,
		accounts:   collections.NewMap(sb, types.AccountsPrefix, "accounts", collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), codec.CollValue[types.AccountsMetadata](cdc)),
	}
}

func (k Keeper) GetAccountsByOwner(ctx context.Context, owner sdk.AccAddress, accountType string) (*types.AccountsMetadata, error) {
	addresses, err := k.accounts.Get(ctx, collections.Join(owner, accountType))

	if err != nil {
		// create new map when key not exist
		if strings.Contains(err.Error(), collections.ErrNotFound.Error()) {
			return &types.AccountsMetadata{
				Addresses: map[string]bool{},
			}, nil
		}

		return nil, err
	}

	return &addresses, nil
}

func (k Keeper) SetAccountsByOwner(ctx context.Context, owner sdk.AccAddress, accountType, address string) error {
	accounts, err := k.GetAccountsByOwner(ctx, owner, accountType)
	if err != nil {
		return err
	}

	exist := accounts.Addresses[address]
	if exist {
		return fmt.Errorf("address %s already exist", address)
	}

	accounts.Addresses[address] = true

	err = k.accounts.Set(ctx, collections.Join(owner, accountType), *accounts)
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
