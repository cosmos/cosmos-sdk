package accounts

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/branch"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/internal/implementation"
)

var (
	errAccountTypeNotFound = errors.New("account type not found")
	// ErrUnauthorized is returned when a message sender is not allowed to perform the operation.
	ErrUnauthorized = errors.New("unauthorized")
)

var (
	// AccountTypeKeyPrefix is the prefix for the account type key.
	AccountTypeKeyPrefix = collections.NewPrefix(0)
	// AccountNumberKey is the key for the account number.
	AccountNumberKey = collections.NewPrefix(1)
	// AccountByNumber is the key for the accounts by number.
	AccountByNumber = collections.NewPrefix(2)
)

// QueryRouter represents a router which can be used to route queries to the correct module.
// It returns the handler given the message name, if multiple handlers are returned, then
// it is up to the caller to choose which one to call.
type QueryRouter interface {
	HybridHandlerByRequestName(name string) []func(ctx context.Context, req, resp implementation.ProtoMsg) error
}

// MsgRouter represents a router which can be used to route messages to the correct module.
type MsgRouter interface {
	HybridHandlerByMsgName(msgName string) func(ctx context.Context, req, resp implementation.ProtoMsg) error
	ResponseNameByRequestName(name string) string
}

// SignerProvider defines an interface used to get the expected sender from a message.
type SignerProvider interface {
	// GetMsgV1Signers returns the signers of the message.
	GetMsgV1Signers(msg gogoproto.Message) ([][]byte, proto.Message, error)
}

// BranchExecutor defines an interface used to execute ops in a branch.
type BranchExecutor = branch.Service

type InterfaceRegistry interface {
	RegisterInterface(name string, iface any, impls ...gogoproto.Message)
	RegisterImplementations(iface any, impls ...gogoproto.Message)
}

func NewKeeper(
	ss store.KVStoreService,
	es event.Service,
	bs BranchExecutor,
	addressCodec address.Codec,
	signerProvider SignerProvider,
	execRouter MsgRouter,
	queryRouter QueryRouter,
	ir InterfaceRegistry,
	accounts ...accountstd.AccountCreatorFunc,
) (Keeper, error) {
	sb := collections.NewSchemaBuilder(ss)
	keeper := Keeper{
		storeService:    ss,
		eventService:    es,
		branchExecutor:  bs,
		addressCodec:    addressCodec,
		signerProvider:  signerProvider,
		msgRouter:       execRouter,
		queryRouter:     queryRouter,
		Schema:          collections.Schema{},
		AccountNumber:   collections.NewSequence(sb, AccountNumberKey, "account_number"),
		AccountsByType:  collections.NewMap(sb, AccountTypeKeyPrefix, "accounts_by_type", collections.BytesKey, collections.StringValue),
		AccountByNumber: collections.NewMap(sb, AccountByNumber, "account_by_number", collections.BytesKey, collections.Uint64Value),
		AccountsState:   collections.NewMap(sb, implementation.AccountStatePrefix, "accounts_state", collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey), collections.BytesValue),
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
	registerToInterfaceRegistry(ir, keeper.accounts)
	return keeper, nil
}

type Keeper struct {
	// deps coming from the runtime
	storeService   store.KVStoreService
	eventService   event.Service
	addressCodec   address.Codec
	branchExecutor BranchExecutor
	msgRouter      MsgRouter
	signerProvider SignerProvider
	queryRouter    QueryRouter

	accounts map[string]implementation.Implementation

	// Schema is the schema for the module.
	Schema collections.Schema
	// AccountNumber is the last global account number.
	AccountNumber collections.Sequence
	// AccountsByType maps account address to their implementation.
	AccountsByType collections.Map[[]byte, string]
	// AccountByNumber maps account number to their address.
	AccountByNumber collections.Map[[]byte, uint64]

	// AccountsState keeps track of the state of each account.
	// NOTE: this is only used for genesis import and export.
	// Account set and get their own state but this helps providing a nice mapping
	// between: (account number, account state key) => account state value.
	AccountsState collections.Map[collections.Pair[uint64, []byte], []byte]
}

// Init creates a new account of the given type.
func (k Keeper) Init(
	ctx context.Context,
	accountType string,
	creator []byte,
	initRequest implementation.ProtoMsg,
) (implementation.ProtoMsg, []byte, error) {
	impl, err := k.getImplementation(accountType)
	if err != nil {
		return nil, nil, err
	}

	// get the next account number
	num, err := k.AccountNumber.Next(ctx)
	if err != nil {
		return nil, nil, err
	}

	// make a new account address
	accountAddr, err := k.makeAddress(num)
	if err != nil {
		return nil, nil, err
	}

	// make the context and init the account
	ctx = k.makeAccountContext(ctx, num, accountAddr, creator, false)
	resp, err := impl.Init(ctx, initRequest)
	if err != nil {
		return nil, nil, err
	}

	// map account address to account type
	if err := k.AccountsByType.Set(ctx, accountAddr, accountType); err != nil {
		return nil, nil, err
	}
	// map account number to account address
	if err := k.AccountByNumber.Set(ctx, accountAddr, num); err != nil {
		return nil, nil, err
	}
	return resp, accountAddr, nil
}

// Execute executes a state transition on the given account.
func (k Keeper) Execute(
	ctx context.Context,
	accountAddr []byte,
	sender []byte,
	execRequest implementation.ProtoMsg,
) (implementation.ProtoMsg, error) {
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

	// get account number
	accountNum, err := k.AccountByNumber.Get(ctx, accountAddr)
	if err != nil {
		return nil, err
	}

	// make the context and execute the account state transition.
	ctx = k.makeAccountContext(ctx, accountNum, accountAddr, sender, false)
	return impl.Execute(ctx, execRequest)
}

// Query queries the given account.
func (k Keeper) Query(
	ctx context.Context,
	accountAddr []byte,
	queryRequest implementation.ProtoMsg,
) (implementation.ProtoMsg, error) {
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

	accountNum, err := k.AccountByNumber.Get(ctx, accountAddr)
	if err != nil {
		return nil, err
	}

	// make the context and execute the account query
	ctx = k.makeAccountContext(ctx, accountNum, accountAddr, nil, true)
	return impl.Query(ctx, queryRequest)
}

func (k Keeper) getImplementation(accountType string) (implementation.Implementation, error) {
	impl, ok := k.accounts[accountType]
	if !ok {
		return implementation.Implementation{}, fmt.Errorf("%w: %s", errAccountTypeNotFound, accountType)
	}
	return impl, nil
}

func (k Keeper) makeAddress(accNum uint64) ([]byte, error) {
	// TODO: better address scheme, ref: https://github.com/cosmos/cosmos-sdk/issues/17516
	addr := sha256.Sum256(append([]byte("x/accounts"), binary.BigEndian.AppendUint64(nil, accNum)...))
	return addr[:], nil
}

// makeAccountContext makes a new context for the given account.
func (k Keeper) makeAccountContext(ctx context.Context, accountNumber uint64, accountAddr, sender []byte, isQuery bool) context.Context {
	// if it's not a query we create a context that allows to do anything.
	if !isQuery {
		return implementation.MakeAccountContext(
			ctx,
			k.storeService,
			accountNumber,
			accountAddr,
			sender,
			k.sendModuleMessage,
			k.sendModuleMessageUntyped,
			k.queryModule,
		)
	}

	// if it's a query we create a context that does not allow to execute modules
	// and does not allow to get the sender.
	return implementation.MakeAccountContext(
		ctx,
		k.storeService,
		accountNumber,
		accountAddr,
		nil,
		func(ctx context.Context, sender []byte, msg, msgResp implementation.ProtoMsg) error {
			return fmt.Errorf("cannot execute in query context")
		},
		func(ctx context.Context, sender []byte, msg implementation.ProtoMsg) (implementation.ProtoMsg, error) {
			return nil, fmt.Errorf("cannot execute in query context")
		},
		k.queryModule,
	)
}

// sendAnyMessages it a helper function that executes untyped codectypes.Any messages
// The messages must all belong to a module.
func (k Keeper) sendAnyMessages(ctx context.Context, sender []byte, anyMessages []*implementation.Any) ([]*implementation.Any, error) {
	anyResponses := make([]*implementation.Any, len(anyMessages))
	for i := range anyMessages {
		msg, err := implementation.UnpackAnyRaw(anyMessages[i])
		if err != nil {
			return nil, err
		}
		resp, err := k.sendModuleMessageUntyped(ctx, sender, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to execute message %d: %s", i, err.Error())
		}
		anyResp, err := implementation.PackAny(resp)
		if err != nil {
			return nil, err
		}
		anyResponses[i] = anyResp
	}
	return anyResponses, nil
}

// sendModuleMessageUntyped can be used to send a message towards a module.
// It should be used when the response type is not known by the caller.
func (k Keeper) sendModuleMessageUntyped(ctx context.Context, sender []byte, msg implementation.ProtoMsg) (implementation.ProtoMsg, error) {
	// we need to fetch the response type from the request message type.
	// this is because the response type is not known.
	respName := k.msgRouter.ResponseNameByRequestName(implementation.MessageName(msg))
	if respName == "" {
		return nil, fmt.Errorf("could not find response type for message %T", msg)
	}
	// get response type
	resp, err := implementation.FindMessageByName(respName)
	if err != nil {
		return nil, err
	}
	// send the message
	return resp, k.sendModuleMessage(ctx, sender, msg, resp)
}

// sendModuleMessage can be used to send a message towards a module. It expects the
// response type to be known by the caller. It will also assert the sender has the right
// is not trying to impersonate another account.
func (k Keeper) sendModuleMessage(ctx context.Context, sender []byte, msg, msgResp implementation.ProtoMsg) error {
	// do sender assertions.
	wantSenders, _, err := k.signerProvider.GetMsgV1Signers(msg)
	if err != nil {
		return fmt.Errorf("cannot get signers: %w", err)
	}
	if len(wantSenders) != 1 {
		return fmt.Errorf("expected only one signer, got %d", len(wantSenders))
	}
	if !bytes.Equal(sender, wantSenders[0]) {
		return fmt.Errorf("%w: sender does not match expected sender", ErrUnauthorized)
	}
	messageName := implementation.MessageName(msg)
	handler := k.msgRouter.HybridHandlerByMsgName(messageName)
	if handler == nil {
		return fmt.Errorf("unknown message: %s", messageName)
	}
	return handler(ctx, msg, msgResp)
}

// queryModule is the entrypoint for an account to query a module.
// It will try to find the query handler for the given query and execute it.
// If multiple query handlers are found, it will return an error.
func (k Keeper) queryModule(ctx context.Context, queryReq, queryResp implementation.ProtoMsg) error {
	queryName := implementation.MessageName(queryReq)
	handlers := k.queryRouter.HybridHandlerByRequestName(queryName)
	if len(handlers) == 0 {
		return fmt.Errorf("unknown query: %s", queryName)
	}
	if len(handlers) > 1 {
		return fmt.Errorf("multiple handlers for query: %s", queryName)
	}
	return handlers[0](ctx, queryReq, queryResp)
}

const msgInterfaceName = "cosmos.accounts.v1.MsgInterface"

// creates a new interface type which is an alias of the proto message interface to avoid conflicts with sdk.Msg
type msgInterface implementation.ProtoMsg

var msgInterfaceType = (*msgInterface)(nil)

// registerToInterfaceRegistry registers all the interfaces of the accounts to the
// global interface registry. This is required for the SDK to correctly decode
// the google.Protobuf.Any used in x/accounts.
func registerToInterfaceRegistry(ir InterfaceRegistry, accMap map[string]implementation.Implementation) {
	ir.RegisterInterface(msgInterfaceName, msgInterfaceType)

	for _, acc := range accMap {
		// register init
		ir.RegisterImplementations(msgInterfaceType, acc.InitHandlerSchema.RequestSchema.New(), acc.InitHandlerSchema.ResponseSchema.New())
		// register exec
		for _, exec := range acc.ExecuteHandlersSchema {
			ir.RegisterImplementations(msgInterfaceType, exec.RequestSchema.New(), exec.ResponseSchema.New())
		}
		// register query
		for _, query := range acc.QueryHandlersSchema {
			ir.RegisterImplementations(msgInterfaceType, query.RequestSchema.New(), query.ResponseSchema.New())
		}
	}
}
