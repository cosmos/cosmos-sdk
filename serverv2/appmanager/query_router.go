package appmanager

import "context"

type QueryHandler = MsgHandler

type QueryRouter struct {
	r MsgRouter
}

func (r QueryRouter) Handle(ctx context.Context, req Type) (resp Type, err error) {
	return r.r.Handle(ctx, req)
}

type QueryRouterBuilder struct {
	b MsgRouterBuilder
}

func (b *QueryRouterBuilder) RegisterHandler(msgType string, handler QueryHandler) {
	b.b.RegisterHandler(msgType, handler)
}

func (b *QueryRouterBuilder) Build() (*QueryRouter, error) {
	msgRouter, err := b.b.Build()
	if err != nil {
		return nil, err
	}
	return &QueryRouter{
		r: *msgRouter,
	}, nil
}
