package wasm

import (
	"bytes"
	"cosmossdk.io/loader/wasm/testdata/test1"
	"encoding/binary"
	"github.com/bytecodealliance/wasmtime-go/v14"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"os"
	"testing"
)

func BenchmarkProto(b *testing.B) {
	m := loadModule(b, "../../rust/target/wasm32-unknown-unknown/release/example_module2.wasm")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		greet := &test1.Greet{
			Name:  "Benchmarker",
			Value: 51,
		}
		bz, err := proto.Marshal(greet)
		require.NoError(b, err)
		inLen := int32(len(bz))
		in, err := m.alloc.Call(m.store, inLen)
		inPtr := in.(int32)
		require.NoError(b, err)
		copy(m.memory.UnsafeData(m.store)[inPtr:inPtr+inLen], bz)
		res, err := m.exec.Call(m.store, inPtr, inLen)
		require.NoError(b, err)
		resI64 := res.(int64)
		outPtr := int32(resI64 & 0xffffffff)
		outLen := int32(resI64 >> 32)
		out := m.memory.UnsafeData(m.store)[outPtr : outPtr+outLen]
		greetRes := &test1.GreetResponse{}
		err = proto.Unmarshal(out, greetRes)
		require.NoError(b, err)
		require.Equal(b, "Hello, Benchmarker! You entered 51", greetRes.Message)
		_, err = m.free.Call(m.store, inPtr, inLen)
		require.NoError(b, err)
		_, err = m.free.Call(m.store, outPtr, outLen)
		require.NoError(b, err)
	}
}

func BenchmarkZeroPB(b *testing.B) {
	m := loadModule(b, "../../rust/target/wasm32-unknown-unknown/release/example_module.wasm")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		err := binary.Write(buf, binary.LittleEndian, uint16(16))
		require.NoError(b, err)
		err = binary.Write(buf, binary.LittleEndian, uint16(11))
		err = binary.Write(buf, binary.LittleEndian, uint32(0)) //padding
		err = binary.Write(buf, binary.LittleEndian, uint64(51))
		require.NoError(b, err)
		_, err = buf.WriteString("Benchmarker")
		require.NoError(b, err)
		bz := buf.Bytes()
		//b.Logf("in: %s", bz)
		n := int32(len(bz))
		in, err := m.alloc.Call(m.store, 0)
		inPtr := in.(int32)
		require.NoError(b, err)
		copy(m.memory.UnsafeData(m.store)[inPtr:inPtr+n], bz)
		binary.LittleEndian.PutUint16(m.memory.UnsafeData(m.store)[inPtr+0x10000-2:inPtr+0x10000], uint16(n))
		res, err := m.exec.Call(m.store, inPtr, 0)
		require.NoError(b, err)
		resI64 := res.(int64)
		outPtr := int32(resI64 & 0xffffffff)
		outLen := int32(binary.LittleEndian.Uint16(m.memory.UnsafeData(m.store)[outPtr+0x10000-2 : outPtr+0x10000]))
		//b.Logf("outPtr: %d, outLen: %d", outPtr, outLen)
		_ = m.memory.UnsafeData(m.store)[outPtr : outPtr+outLen]
		//b.Logf("out: %s", out)
		//_, err = m.free.Call(m.store, inPtr, 0)
		//require.NoError(b, err)
		//_, err = m.free.Call(m.store, outPtr, 0)
		//require.NoError(b, err)
	}
}

type module struct {
	store  *wasmtime.Store
	module *wasmtime.Module
	inst   *wasmtime.Instance
	exec   *wasmtime.Func
	alloc  *wasmtime.Func
	free   *wasmtime.Func
	memory *wasmtime.Memory
}

func loadModule(b testing.TB, path string) module {
	m := module{}
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
