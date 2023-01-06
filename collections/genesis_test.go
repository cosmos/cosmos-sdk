package collections

import (
	"bytes"
	"context"
	"io"
	"testing"

	"cosmossdk.io/core/appmodule"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	f := initFixture(t)
	writers := map[string]*bufCloser{}
	require.NoError(t, f.schema.DefaultGenesis(func(field string) (io.WriteCloser, error) {
		w := newBufCloser(t, "")
		writers[field] = w
		return w, nil
	}))
	require.Len(t, writers, 2)
	require.Equal(t, `[]`, writers["map"].Buffer.String())
	require.Equal(t, `"0"`, writers["item"].Buffer.String())
}

func TestValidateGenesis(t *testing.T) {
	f := initFixture(t)
	require.NoError(t, f.schema.ValidateGenesis(testGenesisSource(t)))
}

func TestImportGenesis(t *testing.T) {
	f := initFixture(t)
	require.NoError(t, f.schema.InitGenesis(f.ctx, testGenesisSource(t)))
	it, err := f.m.Iterate(f.ctx, nil)
	require.NoError(t, err)
	kvs, err := it.KeyValues()
	require.NoError(t, err)
	require.Equal(t, KeyValue[string, uint64]{Key: "abc", Value: 1}, kvs[0])
	require.Equal(t, KeyValue[string, uint64]{Key: "def", Value: 2}, kvs[1])
	x, err := f.i.Get(f.ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(10000), x)
}

func TestExportGenesis(t *testing.T) {
	f := initFixture(t)
	require.NoError(t, f.m.Set(f.ctx, "abc", 1))
	require.NoError(t, f.m.Set(f.ctx, "def", 2))
	require.NoError(t, f.i.Set(f.ctx, 10000))
	writers := map[string]*bufCloser{}
	require.NoError(t, f.schema.ExportGenesis(f.ctx, func(field string) (io.WriteCloser, error) {
		w := newBufCloser(t, "")
		writers[field] = w
		return w, nil
	}))
	require.Len(t, writers, 2)
	require.Equal(t, expectedMapGenesis, writers["map"].Buffer.String())
	require.Equal(t, expectedItemGenesis, writers["item"].Buffer.String())
}

type testFixture struct {
	schema Schema
	ctx    context.Context
	m      Map[string, uint64]
	i      Item[uint64]
}

func initFixture(t *testing.T) *testFixture {
	sk, ctx := deps()
	schemaBuilder := NewSchemaBuilder(sk)
	m := NewMap(schemaBuilder, NewPrefix(1), "map", StringKey, Uint64Value)
	i := NewItem(schemaBuilder, NewPrefix(2), "item", Uint64Value)
	schema, err := schemaBuilder.Build()
	require.NoError(t, err)
	return &testFixture{
		schema: schema,
		ctx:    ctx,
		m:      m,
		i:      i,
	}
}

func testGenesisSource(t *testing.T) appmodule.GenesisSource {
	return func(field string) (io.ReadCloser, error) {
		switch field {
		case "map":
			return newBufCloser(t, expectedMapGenesis), nil
		case "item":
			return newBufCloser(t, expectedItemGenesis), nil
		default:
			return nil, nil
		}
	}
}

const (
	expectedMapGenesis = `[{"key":"abc","value":"1"},
{"key":"def","value":"2"}]`
	expectedItemGenesis = `"10000"`
)

type bufCloser struct {
	*bytes.Buffer
	closed bool
}

func (b *bufCloser) Close() error {
	b.closed = true
	return nil
}

func newBufCloser(t *testing.T, str string) *bufCloser {
	b := &bufCloser{
		Buffer: bytes.NewBufferString(str),
		closed: false,
	}
	// this ensures Close was called by the implementation
	t.Cleanup(func() {
		require.True(t, b.closed)
	})
	return b
}
