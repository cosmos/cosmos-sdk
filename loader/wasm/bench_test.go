package wasm

import (
	"cosmossdk.io/loader/wasm/testdata/test1"
	"encoding/binary"
	"github.com/bytecodealliance/wasmtime-go/v14"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"os"
	"testing"
	"unsafe"
)

func BenchmarkProto(b *testing.B) {
	m := loadWasmModule(b, "../../rust/target/wasm32-unknown-unknown/release/example_module2.wasm")
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

func BenchmarkProtoFFI(b *testing.B) {
	m := LoadFFIModule(b, "../../rust/target/release/libexample_module2.dylib")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		greet := &test1.Greet{
			Name:  "Benchmarker",
			Value: 51,
		}
		bz, err := proto.Marshal(greet)
		require.NoError(b, err)
		out := m.Exec(bz)
		greetRes := &test1.GreetResponse{}
		err = proto.Unmarshal(out, greetRes)
		require.NoError(b, err)
		require.Equal(b, "Hello, Benchmarker! You entered 51", greetRes.Message)
	}
}

type GreetZeroPB struct {
	Name  Str
	Value uint64
}

type Str struct {
	Ptr int16
	Len uint16
}

type GreetResponseZeroPB struct {
	Message Str
}

func BenchmarkZeroPB(b *testing.B) {
	m := loadWasmModule(b, "../../rust/target/wasm32-unknown-unknown/release/example_module.wasm")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bz := sampleZeroPbGreet(b)
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
		out := m.memory.UnsafeData(m.store)[outPtr : outPtr+outLen]
		require.Equal(b, "Hello, Benchmarker! You entered 51", string(out[4:]))
		require.Equal(b, []byte{4, 0, 34, 0}, out[:4])
		//_, err = m.free.Call(m.store, inPtr, 0)
		//require.NoError(b, err)
		//_, err = m.free.Call(m.store, outPtr, 0)
		//require.NoError(b, err)
	}
}

var zeroPbBuf unsafe.Pointer

const sizeOfGreetZeroPB = unsafe.Sizeof(GreetZeroPB{})

func init() {
	writeBuf := make([]byte, 0x20000)
	// find section of writeBuf aligned to 0x10000
	rawPtr := uintptr(unsafe.Pointer(unsafe.SliceData(writeBuf)))
	zeroPbBuf = unsafe.Pointer(rawPtr + (0x10000 - (rawPtr % 0x10000)))
}

func sampleZeroPbGreet(b testing.TB) []byte {
	greet := (*GreetZeroPB)(zeroPbBuf)
	name := "Benchmarker"
	greet.Name.Ptr = int16(sizeOfGreetZeroPB)
	greet.Name.Len = uint16(len(name))
	greet.Value = 51
	strBuf := unsafe.Slice((*byte)(unsafe.Add(zeroPbBuf, sizeOfGreetZeroPB)), len(name))
	copy(strBuf, name)
	return unsafe.Slice((*byte)(zeroPbBuf), 0x10000)
}

func BenchmarkZeroPBFFI(b *testing.B) {
	m := LoadFFIModule(b, "../../rust/target/release/libexample_module.dylib")
	b.ResetTimer()

	//zeroPbBz := []byte{
	//	0x10,
	//	0x0,
	//	0xb,
	//	0x0,
	//	0x0,
	//	0x0,
	//	0x0,
	//	0x0,
	//	0x33,
	//	0x0,
	//	0x0,
	//	0x0,
	//	0x0,
	//	0x0,
	//	0x0,
	//	0x0,
	//	0x42,
	//	0x65,
	//	0x6e,
	//	0x63,
	//	0x68,
	//	0x6d,
	//	0x61,
	//	0x72,
	//	0x6b,
	//	0x65,
	//	0x72,
	//}

	for i := 0; i < b.N; i++ {
		bz := sampleZeroPbGreet(b)
		out := m.Exec(bz)
		require.Equal(b, "Hello, Benchmarker! You entered 51", string(out[4:]))
		require.Equal(b, []byte{4, 0, 34, 0}, out[:4])
		//m.Free(unsafe.Pointer(unsafe.SliceData(out)), len(out))
	}
}

type wasmModule struct {
	store  *wasmtime.Store
	module *wasmtime.Module
	inst   *wasmtime.Instance
	exec   *wasmtime.Func
	alloc  *wasmtime.Func
	free   *wasmtime.Func
	memory *wasmtime.Memory
}

func loadWasmModule(b testing.TB, path string) wasmModule {
	m := wasmModule{}
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
