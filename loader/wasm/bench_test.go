package wasm

import (
	"cosmossdk.io/loader/wasm/testdata/test1"
	"github.com/bytecodealliance/wasmtime-go/v14"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"os"
	"testing"
)

func BenchmarkProto(b *testing.B) {
	store := wasmtime.NewStore(wasmtime.NewEngine())
	bz, err := os.ReadFile("../../rust/target/wasm32-unknown-unknown/release/example_module2.wasm")
	require.NoError(b, err)
	module, err := wasmtime.NewModule(store.Engine, bz)
	require.NoError(b, err)
	inst, err := wasmtime.NewInstance(store, module, nil)
	require.NoError(b, err)
	exec := inst.GetFunc(store, "exec")
	require.NotNil(b, exec)
	alloc := inst.GetFunc(store, "__alloc")
	require.NotNil(b, alloc)
	free := inst.GetFunc(store, "__free")
	require.NotNil(b, free)
	memory := inst.GetExport(store, "memory").Memory()
	require.NotNil(b, memory)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		greet := &test1.Greet{
			Name:  "Benchmarker",
			Value: 51,
		}
		bz, err := proto.Marshal(greet)
		require.NoError(b, err)
		inLen := int32(len(bz))
		in, err := alloc.Call(store, inLen)
		inPtr := in.(int32)
		require.NoError(b, err)
		copy(memory.UnsafeData(store)[inPtr:inPtr+inLen], bz)
		res, err := exec.Call(store, inPtr, inLen)
		require.NoError(b, err)
		resI64 := res.(int64)
		outPtr := int32(resI64 & 0xffffffff)
		outLen := int32(resI64 >> 32)
		out := memory.UnsafeData(store)[outPtr : outPtr+outLen]
		greetRes := &test1.GreetResponse{}
		err = proto.Unmarshal(out, greetRes)
		require.NoError(b, err)
		require.Equal(b, "Hello, Benchmarker! You entered 51", greetRes.Message)
		_, err = free.Call(store, inPtr, inLen)
		_, err = free.Call(store, outPtr, outLen)
		require.NoError(b, err)
	}
}
