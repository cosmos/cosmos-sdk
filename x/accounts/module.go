package accounts

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/tempcore/header"
)

func NewAccounts[H header.Header](storeService store.KVStoreService, headerService header.Service[H]) Accounts[H] {
	sb := collections.NewSchemaBuilder(storeService)

	module := Accounts[H]{
		GlobalAccountNumber: collections.NewSequence(sb, GlobalAccountNumber, "global_account_number"),
		headerService:       headerService,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	module.Schema = schema
	return module
}

type Accounts[H header.Header] struct {
	Schema              collections.Schema
	GlobalAccountNumber collections.Sequence

	headerService header.Service[H]
}
