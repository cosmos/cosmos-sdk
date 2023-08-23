package accounts

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/internal/implementation"
)

var errAccountTypeNotFound = errors.New("account type not found")

var (
	// AccountTypeKeyPrefix is the prefix for the account type key.
	AccountTypeKeyPrefix = collections.NewPrefix(0)
	// AccountNumberKey is the key for the account number.
	AccountNumberKey = collections.NewPrefix(1)
)

func NewModule(ss store.KVStoreService, accounts map[string]implementation.Account) (Module, error) {
	sb := collections.NewSchemaBuilder(ss)
	module := Module{
		storeService:   ss,
		accounts:       map[string]implementation.Implementation{},
		AccountNumber:  collections.NewSequence(sb, AccountNumberKey, "account_number"),
		AccountsByType: collections.NewMap(sb, AccountTypeKeyPrefix, "accounts_by_type", collections.BytesKey, collections.StringValue),
	}

	// make accounts implementation
	for typ, acc := range accounts {
		impl, err := implementation.NewImplementation(acc)
		if err != nil {
			return Module{}, err
		}
		module.accounts[typ] = impl
	}
	schema, err := sb.Build()
	if err != nil {
		return Module{}, err
	}
	module.Schema = schema
	return module, nil
}

type Module struct {
	storeService store.KVStoreService

	accounts map[string]implementation.Implementation

	// Schema is the schema for the module.
	Schema collections.Schema
	// AccountNumber is the last global account number.
	AccountNumber collections.Sequence

	// AccountsByType maps account address to their implementation.
	AccountsByType collections.Map[[]byte, string]
}

// Create creates a new account of the given type.
func (m Module) Create(
	ctx context.Context,
	accountType string,
	creator []byte,
	initRequest any,
) (any, []byte, error) {
	impl, err := m.getImplementation(accountType)
	if err != nil {
		return nil, nil, err
	}

	// make a new account address
	accountAddr, err := m.makeAddress(ctx)
	if err != nil {
		return nil, nil, err
	}

	// make the context and init the account
	ctx = implementation.MakeAccountContext(ctx, m.storeService, accountAddr, creator)
	resp, err := impl.Init(ctx, initRequest)
	if err != nil {
		return nil, nil, err
	}

	// map account address to account type
	if err := m.AccountsByType.Set(ctx, accountAddr, accountType); err != nil {
		return nil, nil, err
	}
	return resp, accountAddr, nil
}

// Execute executes a state transition on the given account.
func (m Module) Execute(
	ctx context.Context,
	accountAddr []byte,
	creator []byte,
	execRequest any,
) (any, error) {
	// get account type
	accountType, err := m.AccountsByType.Get(ctx, accountAddr)
	if err != nil {
		return nil, err
	}

	// get account implementation
	impl, err := m.getImplementation(accountType)
	if err != nil {
		// this means the account was initialized with an implementation
		// that the chain does not know about, in theory should never happen,
		// as it might signal that the app-dev stopped supporting an account type.
		return nil, err
	}

	// make the context and execute the account state transition.
	ctx = implementation.MakeAccountContext(ctx, m.storeService, accountAddr, creator)
	return impl.Execute(ctx, execRequest)
}

// Query queries the given account.
func (m Module) Query(
	ctx context.Context,
	accountAddr []byte,
	queryRequest any,
) (any, error) {
	// get account type
	accountType, err := m.AccountsByType.Get(ctx, accountAddr)
	if err != nil {
		return nil, err
	}

	// get account implementation
	impl, err := m.getImplementation(accountType)
	if err != nil {
		// this means the account was initialized with an implementation
		// that the chain does not know about, in theory should never happen,
		// as it might signal that the app-dev stopped supporting an account type.
		return nil, err
	}

	// make the context and execute the account state transition.
	ctx = implementation.MakeAccountContext(ctx, m.storeService, accountAddr, nil)
	return impl.Query(ctx, queryRequest)
}

func (m Module) getImplementation(accountType string) (implementation.Implementation, error) {
	impl, ok := m.accounts[accountType]
	if !ok {
		return implementation.Implementation{}, fmt.Errorf("%w: %s", errAccountTypeNotFound, accountType)
	}
	return impl, nil
}

func (m Module) makeAddress(ctx context.Context) ([]byte, error) {
	num, err := m.AccountNumber.Next(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: better address scheme, ref: https://github.com/cosmos/cosmos-sdk/issues/17516
	addr := sha256.Sum256(append([]byte("x/accounts"), binary.BigEndian.AppendUint64(nil, num)...))
	return addr[:], nil
}
