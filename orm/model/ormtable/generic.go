package ormtable

import (
	"context"
	"github.com/cosmos/cosmos-sdk/orm/model/ormlist"
	"google.golang.org/protobuf/proto"
)

type GenericTable[Message proto.Message, IndexKey GenericIndexKey] interface {
	Insert(context.Context, ...Message) error
	Update(context.Context, Message) error
	Save(context.Context, Message) error
	Delete(context.Context, ...Message) error
	List(ctx context.Context, prefixKey IndexKey, opts ...ormlist.Option) (GenericIterator[Message], error)
	ListRange(ctx context.Context, from, to IndexKey, opts ...ormlist.Option) (GenericIterator[Message], error)
	DeleteBy(ctx context.Context, prefixKey IndexKey) error
	DeleteRange(ctx context.Context, from, to IndexKey) error
	DynamicTable() Table
	private()
}

type GenericAutoIncrementTable[Message proto.Message, IndexKey GenericIndexKey] interface {
	GenericTable[Message, IndexKey]
	InsertReturningID(context.Context, Message) (newId uint64, err error)
}

type GenericIterator[Message proto.Message] interface {
	Iterator
	Value() (Message, error)
}

type GenericIndexKey interface {
	IndexId() uint32
	KeyValues() []interface{}
}

type genericTable[Message proto.Message, IndexKey GenericIndexKey] struct {
	Table
}

func NewGenericTable[Message proto.Message, IndexKey GenericIndexKey](table Table) GenericTable[Message, IndexKey] {
	return &genericTable[Message, IndexKey]{table}
}

func (g genericTable[M, K]) Insert(ctx context.Context, messages ...M) error {
	for _, message := range messages {
		err := g.Table.Insert(ctx, message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g genericTable[M, K]) Update(ctx context.Context, m M) error {
	return g.Table.Update(ctx, m)
}

func (g genericTable[M, K]) Save(ctx context.Context, m M) error {
	return g.Table.Save(ctx, m)
}

func (g genericTable[M, K]) Delete(ctx context.Context, messages ...M) error {
	for _, message := range messages {
		err := g.Table.Delete(ctx, message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g genericTable[M, K]) List(ctx context.Context, prefixKey K, opts ...ormlist.Option) (GenericIterator[M], error) {
	it, err := g.Table.GetIndexByID(prefixKey.IndexId()).List(ctx, prefixKey.KeyValues(), opts...)
	return &genericIterator[M]{it}, err
}

func (g genericTable[M, K]) ListRange(ctx context.Context, from, to K, opts ...ormlist.Option) (GenericIterator[M], error) {
	it, err := g.Table.GetIndexByID(from.IndexId()).ListRange(ctx, from.KeyValues(), to.KeyValues(), opts...)
	return &genericIterator[M]{it}, err
}

func (g genericTable[M, K]) DeleteBy(ctx context.Context, prefixKey K) error {
	return g.Table.GetIndexByID(prefixKey.IndexId()).DeleteBy(ctx, prefixKey.KeyValues()...)
}

func (g genericTable[M, K]) DeleteRange(ctx context.Context, from, to K) error {
	return g.Table.GetIndexByID(from.IndexId()).DeleteRange(ctx, from.KeyValues(), to.KeyValues())
}

func (g genericTable[M, K]) DynamicTable() Table {
	return g.Table
}

func (g genericTable[M, K]) private() {}

type genericIterator[M proto.Message] struct {
	Iterator
}

func (g genericIterator[M]) Value() (M, error) {
	m, err := g.GetMessage()
	return m.(M), err
}

func NewGenericAutoIncrementTable[Message proto.Message, IndexKey GenericIndexKey](table AutoIncrementTable) GenericAutoIncrementTable[Message, IndexKey] {
	return &genericAutoIncrementTable[Message, IndexKey]{
		GenericTable: &genericTable[Message, IndexKey]{table},
		t:            table.(AutoIncrementTable),
	}
}

type genericAutoIncrementTable[Message proto.Message, IndexKey GenericIndexKey] struct {
	GenericTable[Message, IndexKey]
	t AutoIncrementTable
}

func (g genericAutoIncrementTable[M, K]) InsertReturningID(ctx context.Context, message M) (newId uint64, err error) {
	return g.t.InsertReturningID(ctx, message)
}
