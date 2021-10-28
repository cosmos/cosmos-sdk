package orm

import (
	"context"
	"google.golang.org/protobuf/proto"
)

type client struct {
	orm Server
}

func (c client) Create(ctx context.Context, object proto.Message) error {
	return c.orm.Create(ctx, object)
}

func (c client) Get(ctx context.Context, primaryKey []byte, target proto.Message) error {
	return c.orm.Get(ctx, primaryKey, target)
}

func (c client) Update(ctx context.Context, target proto.Message) error {
	return c.orm.Update(ctx, target)
}

func (c client) Delete(ctx context.Context, object proto.Message) error {
	return c.orm.Delete(ctx, object)
}

func (c client) List(ctx context.Context, object proto.Message, options ListOptions) (objectIterator, error) {
	return c.orm.List(ctx, object, options)
}

type Client interface {
	Create(ctx context.Context, object proto.Message) error
	Get(ctx context.Context, primaryKey []byte, target proto.Message) error
	Update(ctx context.Context, target proto.Message) error
	Delete(ctx context.Context, object proto.Message) error
	List(ctx context.Context, object proto.Message, options ListOptions) (objectIterator, error)
}

func NewClient(orm Server) Client {
	return client{orm: orm}
}
