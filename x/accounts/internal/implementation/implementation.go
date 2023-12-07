package implementation

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
)

// Dependencies are passed to the constructor of a smart account.
type Dependencies struct {
	SchemaBuilder *collections.SchemaBuilder
	AddressCodec  address.Codec
}

// AccountCreatorFunc is a function that creates an account.
type AccountCreatorFunc = func(deps Dependencies) (string, Account, error)

// MakeAccountsMap creates a map of account names to account implementations
// from a list of account creator functions.
func MakeAccountsMap(addressCodec address.Codec, accounts []AccountCreatorFunc) (map[string]Implementation, error) {
	accountsMap := make(map[string]Implementation, len(accounts))
	for _, makeAccount := range accounts {
		stateSchemaBuilder := collections.NewSchemaBuilderFromAccessor(OpenKVStore)
		deps := Dependencies{
			SchemaBuilder: stateSchemaBuilder,
			AddressCodec:  addressCodec,
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
		Init:                  initHandler,
		Execute:               executeHandler,
		Query:                 queryHandler,
		CollectionsSchema:     schema,
		InitHandlerSchema:     ir.schema,
		QueryHandlersSchema:   qr.er.handlersSchema,
		ExecuteHandlersSchema: er.handlersSchema,
	}, nil
}

// Implementation wraps an Account implementer in order to provide a concrete
// and non-generic implementation usable by the x/accounts module.
type Implementation struct {
	// Init defines the initialisation handler for the smart account.
	Init func(ctx context.Context, msg ProtoMsg) (resp ProtoMsg, err error)
	// Execute defines the execution handler for the smart account.
	Execute func(ctx context.Context, msg ProtoMsg) (resp ProtoMsg, err error)
	// Query defines the query handler for the smart account.
	Query func(ctx context.Context, msg ProtoMsg) (resp ProtoMsg, err error)
	// CollectionsSchema represents the state schema.
	CollectionsSchema collections.Schema
	// InitHandlerSchema represents the init handler schema.
	InitHandlerSchema HandlerSchema
	// QueryHandlersSchema is the schema of the query handlers.
	QueryHandlersSchema map[string]HandlerSchema
	// ExecuteHandlersSchema is the schema of the execute handlers.
	ExecuteHandlersSchema map[string]HandlerSchema
}

// MessageSchema defines the schema of a message.
// A message can also define a state schema.
type MessageSchema struct {
	// Name identifies the message name, this must be queriable from some reflection service.
	Name string
	// New is used to create a new message instance for the schema.
	New func() ProtoMsg
}

// HandlerSchema defines the schema of a handler.
type HandlerSchema struct {
	// RequestSchema defines the schema of the request.
	RequestSchema MessageSchema
	// ResponseSchema defines the schema of the response.
	ResponseSchema MessageSchema
}
