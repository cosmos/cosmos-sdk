package testpb

import (
	context "context"
	ormtable "github.com/cosmos/cosmos-sdk/orm/model/ormtable"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)

type TestSchemaStoreAccessor interface {
	Open(context.Context) TestSchemaStore
}

func NewTestSchemaStoreAccessor() (TestSchemaStoreAccessor, error) {
	panic("TODO")
}

type TestSchemaStore interface {
	ExampleTableStore
	ExampleAutoIncrementTableStore
}

type ExampleTableStore interface {
	ExampleTableReader

	CreateExampleTable(exampleTable *ExampleTable) error
	UpdateExampleTable(exampleTable *ExampleTable) error
	SaveExampleTable(exampleTable *ExampleTable) error
	DeleteExampleTable(exampleTable *ExampleTable) error
}

type ExampleTableReader interface {
	HasExampleTable(u32 uint32, i64 int64, str string) (found bool, err error)
	GetExampleTable(u32 uint32, i64 int64, str string) (*ExampleTable, error)
	ListExampleTable(ExampleTableIndexKey) (ExampleTableIterator, error)
}

type ExampleTableIterator struct {
	ormtable.Iterator
}

type ExampleTableIndexKey interface {
	id() uint32
	values() []protoreflect.Value
	exampleTableIndexKey()
}

type ExampleTableU32I64StrIndexKey struct {
}

type ExampleTableU64StrIndexKey struct {
}

type ExampleTableStrU32I64IndexKey struct {
}

type ExampleTableBzStrU32I64IndexKey struct {
}

type ExampleAutoIncrementTableStore interface {
	ExampleAutoIncrementTableReader

	CreateExampleAutoIncrementTable(exampleAutoIncrementTable *ExampleAutoIncrementTable) error
	UpdateExampleAutoIncrementTable(exampleAutoIncrementTable *ExampleAutoIncrementTable) error
	SaveExampleAutoIncrementTable(exampleAutoIncrementTable *ExampleAutoIncrementTable) error
	DeleteExampleAutoIncrementTable(exampleAutoIncrementTable *ExampleAutoIncrementTable) error
}

type ExampleAutoIncrementTableReader interface {
	HasExampleAutoIncrementTable(id uint64) (found bool, err error)
	GetExampleAutoIncrementTable(id uint64) (*ExampleAutoIncrementTable, error)
	ListExampleAutoIncrementTable(ExampleAutoIncrementTableIndexKey) (ExampleAutoIncrementTableIterator, error)
}

type ExampleAutoIncrementTableIterator struct {
	ormtable.Iterator
}

type ExampleAutoIncrementTableIndexKey interface {
	id() uint32
	values() []protoreflect.Value
	exampleAutoIncrementTableIndexKey()
}

type ExampleAutoIncrementTableIdIndexKey struct {
}

type ExampleAutoIncrementTableXIndexKey struct {
}

type testSchemaStore struct {
	exampleTableTable              ormtable.Table
	exampleAutoIncrementTableTable ormtable.Table
	exampleSingletonTable          ormtable.Table
}
