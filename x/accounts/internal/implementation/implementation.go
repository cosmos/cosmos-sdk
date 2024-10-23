package implementation

import (
	"context"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Dependencies are passed to the constructor of a smart account.
type Dependencies struct {
	SchemaBuilder    *collections.SchemaBuilder
	AddressCodec     address.Codec
	Environment      appmodule.Environment
	LegacyStateCodec interface {
		Marshal(gogoproto.Message) ([]byte, error)
		Unmarshal([]byte, gogoproto.Message) error
	}
}

// AccountCreatorFunc is a function that creates an account.
type AccountCreatorFunc = func(deps Dependencies) (string, Account, error)

// MakeAccountsMap creates a map of account names to account implementations
// from a list of account creator functions.
func MakeAccountsMap(
	cdc codec.Codec,
	addressCodec address.Codec,
	env appmodule.Environment,
	accounts []AccountCreatorFunc,
) (map[string]Implementation, error) {
	accountsMap := make(map[string]Implementation, len(accounts))
	for _, makeAccount := range accounts {
		stateSchemaBuilder := collections.NewSchemaBuilderFromAccessor(openKVStore)
		deps := Dependencies{
			SchemaBuilder:    stateSchemaBuilder,
			AddressCodec:     addressCodec,
			Environment:      env,
			LegacyStateCodec: cdc,
		}
		name, accountInterface, err := makeAccount(deps)
		if err != nil {
			return nil, fmt.Errorf("failed to create account %s: %w", name, err)
		}
		if _, ok := accountsMap[name]; ok {
			return nil, fmt.Errorf("account %s is already registered", name)
		}
		impl, err := newImplementation(stateSchemaBuilder, accountInterface)
		if err != nil {
			return nil, fmt.Errorf("failed to create implementation for account %s: %w", name, err)
		}
		accountsMap[name] = impl
	}

	return accountsMap, nil
}

// newImplementation creates a new Implementation instance given an Account implementer.
func newImplementation(schemaBuilder *collections.SchemaBuilder, account Account) (Implementation, error) {
	// make init handler
	ir := NewInitBuilder()
	account.RegisterInitHandler(ir)
	initHandler, err := ir.makeHandler()
	if err != nil {
		return Implementation{}, err
	}

	// make execute handler
	er := NewExecuteBuilder()
	account.RegisterExecuteHandlers(er)
	executeHandler, err := er.makeHandler()
	if err != nil {
		return Implementation{}, err
	}

	// make query handler
	qr := NewQueryBuilder()
	account.RegisterQueryHandlers(qr)
	queryHandler, err := qr.makeHandler()
	if err != nil {
		return Implementation{}, err
	}

	// build schema
	schema, err := schemaBuilder.Build()
	if err != nil {
		return Implementation{}, err
	}
	return Implementation{
		init:                  initHandler,
		execute:               executeHandler,
		query:                 queryHandler,
		collectionsSchema:     schema,
		initHandlerSchema:     ir.schema,
		queryHandlersSchema:   qr.er.handlersSchema,
		executeHandlersSchema: er.handlersSchema,
	}, nil
}

// Implementation wraps an Account implementer in order to provide a concrete
// and non-generic implementation usable by the x/accounts module.
type Implementation struct {
	// init defines the initialisation handler for the smart account.
	init func(ctx context.Context, msg transaction.Msg) (resp transaction.Msg, err error)
	// Execute defines the execution handler for the smart account.
	execute func(ctx context.Context, msg transaction.Msg) (resp transaction.Msg, err error)
	// Query defines the query handler for the smart account.
	query func(ctx context.Context, msg transaction.Msg) (resp transaction.Msg, err error)
	// collectionsSchema represents the state schema.
	collectionsSchema collections.Schema
	// initHandlerSchema represents the init handler schema.
	initHandlerSchema HandlerSchema
	// queryHandlersSchema is the schema of the query handlers.
	queryHandlersSchema map[string]HandlerSchema
	// executeHandlersSchema is the schema of the execute handlers.
	executeHandlersSchema map[string]HandlerSchema
}

func (i Implementation) Init(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
	return i.init(ctx, msg)
}

func (i Implementation) Execute(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
	return i.execute(ctx, msg)
}

func (i Implementation) Query(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
	return i.query(ctx, msg)
}

// HasExec returns true if the account can execute the given msg.
func (i Implementation) HasExec(_ context.Context, m transaction.Msg) bool {
	_, ok := i.executeHandlersSchema[MessageName(m)]
	return ok
}

// HasQuery returns true if the account can execute the given request.
func (i Implementation) HasQuery(_ context.Context, m transaction.Msg) bool {
	_, ok := i.queryHandlersSchema[MessageName(m)]
	return ok
}

func (i Implementation) GetInitHandlerSchema(_ context.Context) (HandlerSchema, error) {
	return i.initHandlerSchema, nil
}

func (i Implementation) GetQueryHandlersSchema(_ context.Context) (map[string]HandlerSchema, error) {
	return i.queryHandlersSchema, nil
}

func (i Implementation) GetExecuteHandlersSchema(_ context.Context) (map[string]HandlerSchema, error) {
	return i.executeHandlersSchema, nil
}

// MessageSchema defines the schema of a message.
// A message can also define a state schema.
type MessageSchema struct {
	// Name identifies the message name, this must be queryable from some reflection service.
	Name string
	// New is used to create a new message instance for the schema.
	New func() transaction.Msg
}

// HandlerSchema defines the schema of a handler.
type HandlerSchema struct {
	// RequestSchema defines the schema of the request.
	RequestSchema MessageSchema
	// ResponseSchema defines the schema of the response.
	ResponseSchema MessageSchema
}

const msgInterfaceName = "cosmos.accounts.v1.MsgInterface"

// creates a new interface type which is an alias of the proto message interface to avoid conflicts with sdk.Msg
type msgInterface transaction.Msg

var msgInterfaceType = (*msgInterface)(nil)

// registerToInterfaceRegistry registers all the interfaces of the accounts to the
// global interface registry. This is required for the SDK to correctly decode
// the google.Protobuf.Any used in x/accounts.
func registerToInterfaceRegistry(ir InterfaceRegistry, accMap map[string]Implementation) {
	ir.RegisterInterface(msgInterfaceName, msgInterfaceType)

	for _, acc := range accMap {
		// register init
		ir.RegisterImplementations(msgInterfaceType, acc.initHandlerSchema.RequestSchema.New(), acc.initHandlerSchema.ResponseSchema.New())
		// register exec
		for _, exec := range acc.executeHandlersSchema {
			ir.RegisterImplementations(msgInterfaceType, exec.RequestSchema.New(), exec.ResponseSchema.New())
		}
		// register query
		for _, query := range acc.queryHandlersSchema {
			ir.RegisterImplementations(msgInterfaceType, query.RequestSchema.New(), query.ResponseSchema.New())
		}
	}
}

type InterfaceRegistry interface {
	RegisterInterface(name string, iface any, impls ...gogoproto.Message)
	RegisterImplementations(iface any, impls ...gogoproto.Message)
}
