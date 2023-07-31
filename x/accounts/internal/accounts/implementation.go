package accounts

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/cosmos/gogoproto/proto"
)

func NewAccountImplementation[
	A Account,
](
	typeName string,
	deps *BuildDependencies,
	constructor func(sb *BuildDependencies) (A, error),
) (Implementation, error) {
	account, err := constructor(deps)
	if err != nil {
		return Implementation{}, err
	}
	executeRouter := &ExecuteRouter{}
	err = account.RegisterExecuteHandlers(executeRouter)
	if err != nil {
		return Implementation{}, err
	}

	queryRouter := &QueryRouter{}
	err = account.RegisterQueryHandlers(queryRouter)
	if err != nil {
		return Implementation{}, err
	}

	initRouter := &InitRouter{}
	err = account.RegisterInitHandler(initRouter)

	// build schema
	schema, err := deps.SchemaBuilder.Build()
	if err != nil {
		return Implementation{}, err
	}
	return Implementation{
		StateSchema: schema,
		Execute:     executeRouter.Handler(),
		Query:       queryRouter.Handler(),
		Init:        initRouter.Handler(),
		Type:        typeName,
	}, nil
}

// Implementation represents the implementation of the Accounts module.
type Implementation struct {
	StateSchema collections.Schema
	Execute     func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Query       func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Init        func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Type        string
}
