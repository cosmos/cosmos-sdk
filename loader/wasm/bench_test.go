package wasm

import (
	"cosmossdk.io/loader/wasm/testdata/test1"
	"encoding/binary"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"
	"unsafe"
)

func BenchmarkProtoWasm(b *testing.B) {
	m := LoadWasmModule(b, "../../rust/target/wasm32-unknown-unknown/release/example_module2.wasm")
	protoWasmRound(b, m, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		protoWasmRound(b, m, false)
	}
}

func protoWasmRound(b testing.TB, m WasmModule, check bool) {
	greet := &test1.Greet{
		Name:  "Benchmarker",
		Value: 51,
	}
	bz, err := proto.Marshal(greet)
	if check {
		require.NoError(b, err)
	}
	inLen := int32(len(bz))
	inPtr, err := m.Alloc(inLen)
	if check {
		require.NoError(b, err)
	}
	copy(m.memory.UnsafeData(m.store)[inPtr:inPtr+inLen], bz)
	outPtr, outLen, err := m.Exec(inPtr, inLen)
	if check {
		require.NoError(b, err)
	}
	out := m.memory.UnsafeData(m.store)[outPtr : outPtr+outLen]
	greetRes := &test1.GreetResponse{}
	err = proto.Unmarshal(out, greetRes)
	if check {
		require.NoError(b, err)
		require.Equal(b, "Hello, Benchmarker! You entered 51", greetRes.Message)
	}
	err = m.Free(inPtr, inLen)
	if check {
		require.NoError(b, err)
	}
	err = m.Free(outPtr, outLen)
	if check {
		require.NoError(b, err)
	}
}

func BenchmarkProtoFFI(b *testing.B) {
	m := LoadFFIModule(b, "../../rust/target/release/libexample_module2.dylib")
	protoFFIRound(b, m, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		protoFFIRound(b, m, false)
	}
}

func protoFFIRound(b testing.TB, m FFIModule, check bool) {
	greet := &test1.Greet{
		Name:  "Benchmarker",
		Value: 51,
	}
	bz, err := proto.Marshal(greet)
	if check {
		require.NoError(b, err)
	}
	out := m.Exec(bz)
	greetRes := &test1.GreetResponse{}
	err = proto.Unmarshal(out, greetRes)
	if check {
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

func BenchmarkZeroPBWasm(b *testing.B) {
	m := LoadWasmModule(b, "../../rust/target/wasm32-unknown-unknown/release/example_module.wasm")
	zeroPBWasmRound(b, m, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		zeroPBWasmRound(b, m, false)
	}

	b.StopTimer()
	allocations := m.inst.GetFunc(m.store, "allocations")
	require.NotNil(b, allocations)
	res, err := allocations.Call(m.store)
	require.NoError(b, err)
	require.Equal(b, int32(2), res.(int32))
}

func zeroPBWasmRound(b testing.TB, m WasmModule, check bool) {
	bz, n := sampleZeroPbGreet(b)
	inPtr, err := m.Alloc(0x10000)
	if check {
		require.NoError(b, err)
	}
	inMem := m.memory.UnsafeData(m.store)
	copy(inMem[inPtr:inPtr+n], bz)
	// write extent pointer
	binary.LittleEndian.PutUint16(inMem[inPtr+0x10000-2:inPtr+0x10000], uint16(n))
	outPtr, _, err := m.Exec(inPtr, n)
	outMem := m.memory.UnsafeData(m.store)
	outLen := int32(binary.LittleEndian.Uint16(outMem[outPtr+0x10000-2 : outPtr+0x10000]))
	out := outMem[outPtr : outPtr+outLen]
	if check {
		require.NoError(b, err)
		require.Equal(b, "Hello, Benchmarker! You entered 51", string(out[4:]))
		require.Equal(b, []byte{4, 0, 34, 0}, out[:4])
		require.Equal(b, outLen, int32(4+34))
	}
	err = m.Free(inPtr, 0x10000)
	if check {
		require.NoError(b, err)
	}
	err = m.Free(outPtr, outLen)
	if check {
		require.NoError(b, err)
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

func sampleZeroPbGreet(b testing.TB) ([]byte, int32) {
	greet := (*GreetZeroPB)(zeroPbBuf)
	name := "Benchmarker"
	greet.Name.Ptr = int16(sizeOfGreetZeroPB)
	greet.Name.Len = uint16(len(name))
	greet.Value = 51
	n := len(name)
	strBuf := unsafe.Slice((*byte)(unsafe.Add(zeroPbBuf, sizeOfGreetZeroPB)), n)
	copy(strBuf, name)
	n += int(sizeOfGreetZeroPB)
	return unsafe.Slice((*byte)(zeroPbBuf), 0x10000), int32(n)
}

func BenchmarkZeroPBFFI(b *testing.B) {
	m := LoadFFIModule(b, "../../rust/target/release/libexample_module.dylib")
	zeroPBFFIRound(b, m, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		zeroPBFFIRound(b, m, false)
	}
}

func zeroPBFFIRound(b testing.TB, m FFIModule, check bool) {
	bz, n := sampleZeroPbGreet(b)
	binary.LittleEndian.PutUint16(bz[0x10000-2:0x10000], uint16(n))
	out := m.Exec(bz)
	if check {
		require.Equal(b, []byte{4, 0, 34, 0}, out[:4])
		require.Equal(b, "Hello, Benchmarker! You entered 51", string(out[4:]))
	}
}
