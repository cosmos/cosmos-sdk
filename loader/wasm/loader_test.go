package wasm

import (
	"github.com/bytecodealliance/wasmtime-go/v14"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test1(t *testing.T) {
	store := wasmtime.NewStore(wasmtime.NewEngine())
	bz, err := os.ReadFile("../../rust/target/wasm32-unknown-unknown/release/example_module.wasm")
	require.NoError(t, err)
	module, err := wasmtime.NewModule(store.Engine, bz)
	for _, importType := range module.Imports() {
		t.Logf("import: %s %s %+v", importType.Module(), *importType.Name(), importType.Type())
	}
	require.NoError(t, err)
	_, err = wasmtime.NewInstance(store, module, nil)
	require.NoError(t, err)
}
