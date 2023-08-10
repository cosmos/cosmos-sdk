package accounts

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	internalaccounts "cosmossdk.io/x/accounts/internal/accounts"
	"cosmossdk.io/x/accounts/sdk"
	"cosmossdk.io/x/accounts/tempcore/header"
)

func AddAccount[A sdk.Account](
	name string,
	constructor func(deps *sdk.BuildDependencies) (A, error),
) func(deps *sdk.BuildDependencies) (internalaccounts.Implementation, error) {
	return func(deps *sdk.BuildDependencies) (internalaccounts.Implementation, error) {
		return internalaccounts.NewAccountImplementation(name, deps, constructor)
	}
}

func NewAccounts[
	H header.Header,
](
	addressCodec address.Codec,
	storeService store.KVStoreService,
	headerService header.Service[H],
	accounts ...func(bd *sdk.BuildDependencies) (internalaccounts.Implementation, error),
) (Accounts[H], error) {
	sb := collections.NewSchemaBuilder(storeService)

	module := Accounts[H]{
		Schema:              collections.Schema{},
		GlobalAccountNumber: collections.NewSequence(sb, GlobalAccountNumberPrefix, "global_account_number"),
		AccountsByType:      collections.NewMap(sb, AccountsTypePrefix, "accounts_by_type", collections.BytesKey, collections.StringValue),
		AccountsState:       collections.NewMap(sb, AccountsStatePrefix, "accounts_state", collections.BytesKey, collections.BytesValue),
		headerService:       headerService,
		storeService:        storeService,
		addressCodec:        addressCodec,
		accounts:            map[string]internalaccounts.Implementation{},
	}

	schema, err := sb.Build()
	if err != nil {
		return Accounts[H]{}, err
	}
	module.Schema = schema

	err = module.registerAccounts(accounts...)
	if err != nil {
		return Accounts[H]{}, err
	}
	return module, nil
}

type Accounts[H header.Header] struct {
	Schema collections.Schema

	GlobalAccountNumber collections.Sequence
	AccountsByType      collections.Map[[]byte, string] // map an account address to its type.
	AccountsState       collections.Map[[]byte, []byte] // this is never written to by the accounts module, useful for single account state dump.

	headerService header.Service[H]
	storeService  store.KVStoreService
	addressCodec  address.Codec

	accounts map[string]internalaccounts.Implementation // maps an account implementation by its type.
}

func (a Accounts[H]) registerAccounts(constructors ...func(deps *sdk.BuildDependencies) (internalaccounts.Implementation, error)) error {
	for _, constructor := range constructors {
		deps := internalaccounts.MakeBuildDependencies(a.Execute, a.Query)
		accountImpl, err := constructor(deps)
		if err != nil {
			return err
		}

		typ := accountImpl.Type
		if _, ok := a.accounts[typ]; ok {
			return fmt.Errorf("account type %s already registered", typ)
		}
		a.accounts[typ] = accountImpl
	}

	return nil
}
