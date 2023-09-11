package accounts

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/internal/implementation"
	"google.golang.org/protobuf/proto"
)

var errAccountTypeNotFound = errors.New("account type not found")

var (
	// AccountTypeKeyPrefix is the prefix for the account type key.
	AccountTypeKeyPrefix = collections.NewPrefix(0)
	// AccountNumberKey is the key for the account number.
	AccountNumberKey = collections.NewPrefix(1)
)

func NewKeeper(
	ss store.KVStoreService,
	addressCodec address.Codec,
	getMsgSenderFunc func(msg proto.Message) ([]byte, error),
	execModuleFunc func(ctx context.Context, msg proto.Message) (proto.Message, error),
	queryModuleFunc func(ctx context.Context, msg proto.Message) (proto.Message, error),
	accounts map[string]implementation.Account,
) (Keeper, error) {
	sb := collections.NewSchemaBuilder(ss)
	keeper := Keeper{
		storeService:    ss,
		addressCodec:    addressCodec,
		getSenderFunc:   getMsgSenderFunc,
		execModuleFunc:  execModuleFunc,
		queryModuleFunc: queryModuleFunc,
		accounts:        map[string]implementation.Implementation{},
		Schema:          collections.Schema{},
		AccountNumber:   collections.NewSequence(sb, AccountNumberKey, "account_number"),
		AccountsByType:  collections.NewMap(sb, AccountTypeKeyPrefix, "accounts_by_type", collections.BytesKey, collections.StringValue),
	}

	// make accounts implementation
	for typ, acc := range accounts {
		impl, err := implementation.NewImplementation(acc)
		if err != nil {
			return Keeper{}, err
		}
		keeper.accounts[typ] = impl
	}
	schema, err := sb.Build()
	if err != nil {
		return Keeper{}, err
	}
	keeper.Schema = schema
	return keeper, nil
}

type Keeper struct {
	// deps coming from the runtime
	storeService    store.KVStoreService
	addressCodec    address.Codec
	getSenderFunc   func(msg proto.Message) ([]byte, error)
	execModuleFunc  func(ctx context.Context, msg proto.Message) (proto.Message, error)
	queryModuleFunc func(ctx context.Context, msg proto.Message) (proto.Message, error)

	accounts map[string]implementation.Implementation

	// Schema is the schema for the module.
	Schema collections.Schema
	// AccountNumber is the last global account number.
	AccountNumber collections.Sequence
	// AccountsByType maps account address to their implementation.
	AccountsByType collections.Map[[]byte, string]
}

// Init creates a new account of the given type.
func (k Keeper) Init(
	ctx context.Context,
	accountType string,
	creator []byte,
	initRequest any,
) (any, []byte, error) {
	impl, err := k.getImplementation(accountType)
	if err != nil {
		return nil, nil, err
	}

	// make a new account address
	accountAddr, err := k.makeAddress(ctx)
	if err != nil {
		return nil, nil, err
	}

	// make the context and init the account
	ctx = implementation.MakeAccountContext(ctx, k.storeService, accountAddr, creator, k.getSenderFunc, k.execModuleFunc, k.queryModuleFunc)
	resp, err := impl.Init(ctx, initRequest)
	if err != nil {
		return nil, nil, err
	}

	// map account address to account type
	if err := k.AccountsByType.Set(ctx, accountAddr, accountType); err != nil {
		return nil, nil, err
	}
	return resp, accountAddr, nil
}

// Execute executes a state transition on the given account.
func (k Keeper) Execute(
	ctx context.Context,
	accountAddr []byte,
	sender []byte,
	execRequest any,
) (any, error) {
	// get account type
	accountType, err := k.AccountsByType.Get(ctx, accountAddr)
	if err != nil {
		return nil, err
	}

	// get account implementation
	impl, err := k.getImplementation(accountType)
	if err != nil {
		// this means the account was initialized with an implementation
		// that the chain does not know about, in theory should never happen,
		// as it might signal that the app-dev stopped supporting an account type.
		return nil, err
	}

	// make the context and execute the account state transition.
	ctx = implementation.MakeAccountContext(ctx, k.storeService, accountAddr, sender, k.getSenderFunc, k.execModuleFunc, k.queryModuleFunc)
	return impl.Execute(ctx, execRequest)
}

// Query queries the given account.
func (k Keeper) Query(
	ctx context.Context,
	accountAddr []byte,
	queryRequest any,
) (any, error) {
	// get account type
	accountType, err := k.AccountsByType.Get(ctx, accountAddr)
	if err != nil {
		return nil, err
	}

	// get account implementation
	impl, err := k.getImplementation(accountType)
	if err != nil {
		// this means the account was initialized with an implementation
		// that the chain does not know about, in theory should never happen,
		// as it might signal that the app-dev stopped supporting an account type.
		return nil, err
	}

	// make the context and execute the account state transition.
	// we do not pass in sender as a query does not have a sender.
	// we do not pass in execModuleFunc as a query does not execute any module.
	// we do not pass in getSenderFunc as a query does not have a sender.
	ctx = implementation.MakeAccountContext(ctx, k.storeService, accountAddr, nil, func(_ proto.Message) ([]byte, error) {
		return nil, fmt.Errorf("cannot get sender from query")
	}, func(ctx context.Context, _ proto.Message) (proto.Message, error) {
		return nil, fmt.Errorf("cannot execute module from query")
	}, k.queryModuleFunc)
	return impl.Query(ctx, queryRequest)
}

func (k Keeper) getImplementation(accountType string) (implementation.Implementation, error) {
	impl, ok := k.accounts[accountType]
	if !ok {
		return implementation.Implementation{}, fmt.Errorf("%w: %s", errAccountTypeNotFound, accountType)
	}
	return impl, nil
}

func (k Keeper) makeAddress(ctx context.Context) ([]byte, error) {
	num, err := k.AccountNumber.Next(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: better address scheme, ref: https://github.com/cosmos/cosmos-sdk/issues/17516
	addr := sha256.Sum256(append([]byte("x/accounts"), binary.BigEndian.AppendUint64(nil, num)...))
	return addr[:], nil
}
