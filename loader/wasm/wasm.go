package wasm

import (
	"github.com/bytecodealliance/wasmtime-go/v14"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

type WasmModule struct {
	store  *wasmtime.Store
	module *wasmtime.Module
	inst   *wasmtime.Instance
	exec   *wasmtime.Func
	alloc  *wasmtime.Func
	free   *wasmtime.Func
	memory *wasmtime.Memory
}

func LoadWasmModule(b testing.TB, path string) WasmModule {
	m := WasmModule{}
	m.store = wasmtime.NewStore(wasmtime.NewEngine())
	bz, err := os.ReadFile(path)
	require.NoError(b, err)
	m.module, err = wasmtime.NewModule(m.store.Engine, bz)
	require.NoError(b, err)
	m.inst, err = wasmtime.NewInstance(m.store, m.module, nil)
	require.NoError(b, err)
	m.exec = m.inst.GetFunc(m.store, "exec")
	require.NotNil(b, m.exec)
	m.alloc = m.inst.GetFunc(m.store, "__alloc")
	require.NotNil(b, m.alloc)
	m.free = m.inst.GetFunc(m.store, "__free")
	require.NotNil(b, m.free)
	m.memory = m.inst.GetExport(m.store, "memory").Memory()
	require.NotNil(b, m.memory)
	return m
}

func (w WasmModule) Alloc(n int32) (int32, error) {
	res, err := w.alloc.Call(w.store, n)
	return res.(int32), err
}

func (w WasmModule) Free(ptr int32, n int32) error {
	_, err := w.free.Call(w.store, ptr, n)
	return err
}

func (w WasmModule) Exec(inPtr int32, inLen int32) (outPtr int32, outLen int32, err error) {
	res, err := w.exec.Call(w.store, inPtr, inLen)
	resI64 := res.(int64)
	outPtr = int32(resI64 & 0xffffffff)
	outLen = int32(resI64 >> 32)
	return
}
