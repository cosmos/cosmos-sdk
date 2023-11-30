package wasm

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/bytecodealliance/wasmtime-go/v14"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type WasmModule struct {
	store       *wasmtime.Store
	module      *wasmtime.Module
	inst        *wasmtime.Instance
	exec        *wasmtime.Func
	execMsgSend *wasmtime.Func
	alloc       *wasmtime.Func
	free        *wasmtime.Func
	memory      *wasmtime.Memory
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
	require.NoError(b, err)
	m.execMsgSend = m.inst.GetFunc(m.store, "exec_msg_send")
	require.NotNil(b, m.execMsgSend)
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

func (w WasmModule) Free(ptr int32, n int32) {
	_, err := w.free.Call(w.store, ptr, n)
	if err != nil {
		panic(err)
	}
}

func (w WasmModule) Exec(inPtr int32, inLen int32) (outPtr int32, outLen int32) {
	return w.doExec(w.exec, inPtr, inLen)
}

func (w WasmModule) ExecMsgSend(inPtr int32, inLen int32) (outPtr int32, outLen int32) {
	return w.doExec(w.execMsgSend, inPtr, inLen)
}

func (w WasmModule) doExec(f *wasmtime.Func, inPtr int32, inLen int32) (outPtr int32, outLen int32) {
	res, err := f.Call(w.store, inPtr, inLen)
	if err != nil {
		panic(err)
	}
	resI64 := res.(int64)
	outPtr = int32(resI64 & 0xffffffff)
	outLen = int32(resI64 >> 32)
	return
}

func (w WasmModule) WriteProto(msg proto.Message) (int32, int32) {
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	inLen := int32(len(bz))
	inPtr, err := w.Alloc(inLen)
	if err != nil {
		panic(err)
	}
	copy(w.memory.UnsafeData(w.store)[inPtr:inPtr+inLen], bz)
	return inPtr, inLen
}

func (w WasmModule) WriteZeroPB(bz []byte, n int32) int32 {
	inPtr, err := w.Alloc(0x10000)
	if err != nil {
		panic(err)
	}
	inMem := w.memory.UnsafeData(w.store)
	copy(inMem[inPtr:inPtr+n], bz)
	// write extent pointer
	binary.LittleEndian.PutUint16(inMem[inPtr+0x10000-2:inPtr+0x10000], uint16(n))
	return inPtr
}

func (w WasmModule) ReadZeroPBOutPtr(outPtr int32) ([]byte, int32) {
	outMem := w.memory.UnsafeData(w.store)
	outLen := int32(binary.LittleEndian.Uint16(outMem[outPtr+0x10000-2 : outPtr+0x10000]))
	return outMem[outPtr : outPtr+outLen], outLen
}

func (w WasmModule) ReadProtoOut(outPtr, outLen int32, msg proto.Message) {
	out := w.memory.UnsafeData(w.store)[outPtr : outPtr+outLen]
	err := proto.Unmarshal(out, msg)
	if err != nil {
		panic(err)
	}
}
