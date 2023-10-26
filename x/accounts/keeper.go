package accounts

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	"cosmossdk.io/core/event"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var errAccountTypeNotFound = errors.New("account type not found")

var (
	// AccountTypeKeyPrefix is the prefix for the account type key.
	AccountTypeKeyPrefix = collections.NewPrefix(0)
	// AccountNumberKey is the key for the account number.
	AccountNumberKey = collections.NewPrefix(1)
)

// QueryRouter represents a router which can be used to route queries to the correct module.
// It returns the handler given the message name, if multiple handlers are returned, then
// it is up to the caller to choose which one to call.
type QueryRouter interface {
	HybridHandlerByRequestName(name string) []func(ctx context.Context, req, resp protoiface.MessageV1) error
}

// MsgRouter represents a router which can be used to route messages to the correct module.
type MsgRouter interface {
	HybridHandlerByMsgName(msgName string) func(ctx context.Context, req, resp protoiface.MessageV1) error
}

// SignerProvider defines an interface used to get the expected sender from a message.
type SignerProvider interface {
	// GetSigners returns the signers of the message.
	GetSigners(msg proto.Message) ([][]byte, error)
}

func NewKeeper(
	ss store.KVStoreService,
	es event.Service,
	addressCodec address.Codec,
	signerProvider SignerProvider,
	execRouter MsgRouter,
	queryRouter QueryRouter,
	accounts ...accountstd.AccountCreatorFunc,
) (Keeper, error) {
	sb := collections.NewSchemaBuilder(ss)
	keeper := Keeper{
		storeService: ss,
		eventService: es,
		addressCodec: addressCodec,
		getSenderFunc: func(msg proto.Message) ([]byte, error) {
			signers, err := signerProvider.GetSigners(msg)
			if err != nil {
				return nil, err
			}
			if len(signers) != 1 {
				return nil, fmt.Errorf("expected 1 signer, got %d", len(signers))
			}
			return signers[0], nil
		},
		execModuleFunc: func(ctx context.Context, msg, msgResp protoiface.MessageV1) error {
			name := getMessageName(msg)
			handler := execRouter.HybridHandlerByMsgName(name)
			if handler == nil {
				return fmt.Errorf("no handler found for message %s", name)
			}
			return handler(ctx, msg, msgResp)
		},
		queryModuleFunc: func(ctx context.Context, req, resp protoiface.MessageV1) error {
			name := getMessageName(req)
			handlers := queryRouter.HybridHandlerByRequestName(name)
			if len(handlers) == 0 {
				return fmt.Errorf("no handler found for query request %s", name)
			}
			if len(handlers) > 1 {
				return fmt.Errorf("multiple handlers found for query request %s", name)
			}
			return handlers[0](ctx, req, resp)
		},
		AccountNumber:  collections.NewSequence(sb, AccountNumberKey, "account_number"),
		AccountsByType: collections.NewMap(sb, AccountTypeKeyPrefix, "accounts_by_type", collections.BytesKey, collections.StringValue),
		AccountsState:  collections.NewMap(sb, implementation.AccountStatePrefix, "accounts_state", collections.BytesKey, collections.BytesValue),
	}

	schema, err := sb.Build()
	if err != nil {
		return Keeper{}, err
	}
	keeper.Schema = schema
	keeper.accounts, err = implementation.MakeAccountsMap(keeper.addressCodec, accounts)
	if err != nil {
		return Keeper{}, err
	}
	return keeper, nil
}

type Keeper struct {
	// deps coming from the runtime
	storeService    store.KVStoreService
	eventService    event.Service
	addressCodec    address.Codec
	getSenderFunc   func(msg proto.Message) ([]byte, error)
	execModuleFunc  implementation.ModuleExecFunc
	queryModuleFunc implementation.ModuleQueryFunc

	accounts map[string]implementation.Implementation

	// Schema is the schema for the module.
	Schema collections.Schema
	// AccountNumber is the last global account number.
	AccountNumber collections.Sequence
	// AccountsByType maps account address to their implementation.
	AccountsByType collections.Map[[]byte, string]

	// AccountsState keeps track of the state of each account.
	// NOTE: this is only used for genesis import and export.
	// Contracts set and get their own state but this helps providing a nice mapping
	// between: (account address, account state key) => account state value.
	AccountsState collections.Map[[]byte, []byte]
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
	ctx = k.makeAccountContext(ctx, accountAddr, creator, false)
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
	ctx = k.makeAccountContext(ctx, accountAddr, sender, false)
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

	// make the context and execute the account query
	ctx = k.makeAccountContext(ctx, accountAddr, nil, true)
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

// makeAccountContext makes a new context for the given account.
func (k Keeper) makeAccountContext(ctx context.Context, accountAddr, sender []byte, isQuery bool) context.Context {
	// if it's not a query we create a context that allows to do anything.
	if !isQuery {
		return implementation.MakeAccountContext(
			ctx,
			k.storeService,
			accountAddr,
			sender,
			k.getSenderFunc,
			k.execModuleFunc,
			k.queryModuleFunc,
		)
	}

	// if it's a query we create a context that does not allow to execute modules
	// and does not allow to get the sender.
	return implementation.MakeAccountContext(
		ctx,
		k.storeService,
		accountAddr,
		nil,
		func(_ proto.Message) ([]byte, error) {
			return nil, fmt.Errorf("cannot get sender from query")
		},
		func(ctx context.Context, msg, msgResp protoiface.MessageV1) error {
			return fmt.Errorf("cannot execute module from a query execution context")
		},
		k.queryModuleFunc,
	)
}

func getMessageName(msg protoiface.MessageV1) string {
	return codectypes.MsgTypeURL(msg)[1:]
}
