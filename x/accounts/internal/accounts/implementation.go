package accounts

import (
	"context"

	"google.golang.org/protobuf/proto"
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
	account.RegisterExecuteHandlers(executeRouter)
	queryRouter := &QueryRouter{}
	account.RegisterQueryHandlers(queryRouter)
	initRouter := &InitRouter{}
	account.RegisterInitHandler(initRouter)
	execHandler, err := executeRouter.Handler()
	if err != nil {
		return Implementation{}, err
	}
	queryHandler, err := queryRouter.Handler()
	if err != nil {
		return Implementation{}, err
	}

	// build schema
	schemas, err := NewSchemas(deps.SchemaBuilder, initRouter, executeRouter, queryRouter)
	if err != nil {
		return Implementation{}, err
	}

	return Implementation{
		Schemas: schemas,
		Execute: execHandler,
		Query:   queryHandler,
		Init:    initRouter.Handler(),
		Type:    typeName,
	}, nil
}

// Implementation represents the implementation of the Accounts module.
type Implementation struct {
	Schemas *Schemas

	Execute func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Query   func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Init    func(ctx context.Context, msg proto.Message) (proto.Message, error)
	Type    string
}
