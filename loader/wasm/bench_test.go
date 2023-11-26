package wasm

import (
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/loader/wasm/testdata/test1"
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
	outPtr, outLen := m.Exec(inPtr, inLen)
	out := m.memory.UnsafeData(m.store)[outPtr : outPtr+outLen]
	greetRes := &test1.GreetResponse{}
	err = proto.Unmarshal(out, greetRes)
	if check {
		require.NoError(b, err)
		require.Equal(b, "Hello, Benchmarker! You entered 51", greetRes.Message)
	}
	m.Free(inPtr, inLen)
	m.Free(outPtr, outLen)
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

	checkWasmAllocations(b, m)
}

func checkWasmAllocations(b *testing.B, m WasmModule) {
	b.StopTimer()
	allocations := m.inst.GetFunc(m.store, "allocations")
	require.NotNil(b, allocations)
	res, err := allocations.Call(m.store)
	require.NoError(b, err)
	require.Equal(b, int32(2), res.(int32))
}

func zeroPBWasmRound(b testing.TB, m WasmModule, check bool) {
	bz, n := sampleZeroPbGreet(b)
	inPtr := m.WriteZeroPB(b, bz, n, check)
	outPtr, _ := m.Exec(inPtr, n)
	out, outLen := m.ReadZeroPBOutPtr(outPtr)
	if check {
		require.Equal(b, "Hello, Benchmarker! You entered 51", string(out[4:]))
		require.Equal(b, []byte{4, 0, 34, 0}, out[:4])
		require.Equal(b, outLen, int32(4+34))
	}
	m.Free(inPtr, 0x10000)
	m.Free(outPtr, outLen)
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

	b.StopTimer()
	allocations := m.Allocations()
	require.Equal(b, int32(1), allocations)
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

func BenchmarkZeroPBWasmMsgSend(b *testing.B) {
	m := LoadWasmModule(b, "../../rust/target/wasm32-unknown-unknown/release/example_module.wasm")
	zeroPBWasmMsgSendRound(b, m, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		zeroPBWasmMsgSendRound(b, m, false)
	}

	checkWasmAllocations(b, m)
}

func zeroPBWasmMsgSendRound(b *testing.B, m WasmModule, check bool) {
	inPtr := m.WriteZeroPB(b, testMsgSendZeroPbBz, int32(len(testMsgSendZeroPbBz)), check)
	outPtr, _ := m.ExecMsgSend(inPtr, int32(len(testMsgSendZeroPbBz)))
	out, outLen := m.ReadZeroPBOutPtr(outPtr)
	if check {
		require.Equal(b, "cosmos1huydeevpz37sd9snkgul6070mstupukw00xkw9 sent to cosmos1xy4yqngt0nlkdcenxymg8tenrghmek4nmqm28k:  1234567 uatom  7654321 foo", string(out[4:]))
		require.Equal(b, []byte{0x77, 0, 0xd, 0}, out[:4])
		require.Equal(b, int32(132), outLen)
	}
	m.Free(inPtr, 0x10000)
	m.Free(outPtr, 0x10000)
}

var testMsgSendZeroPbBz = []byte{
	0xc,
	0x0,
	0x2d,
	0x0,
	0x35,
	0x0,
	0x2d,
	0x0,
	0x5e,
	0x0,
	0x2,
	0x0,
	0x63,
	0x6f,
	0x73,
	0x6d,
	0x6f,
	0x73,
	0x31,
	0x68,
	0x75,
	0x79,
	0x64,
	0x65,
	0x65,
	0x76,
	0x70,
	0x7a,
	0x33,
	0x37,
	0x73,
	0x64,
	0x39,
	0x73,
	0x6e,
	0x6b,
	0x67,
	0x75,
	0x6c,
	0x36,
	0x30,
	0x37,
	0x30,
	0x6d,
	0x73,
	0x74,
	0x75,
	0x70,
	0x75,
	0x6b,
	0x77,
	0x30,
	0x30,
	0x78,
	0x6b,
	0x77,
	0x39,
	0x63,
	0x6f,
	0x73,
	0x6d,
	0x6f,
	0x73,
	0x31,
	0x78,
	0x79,
	0x34,
	0x79,
	0x71,
	0x6e,
	0x67,
	0x74,
	0x30,
	0x6e,
	0x6c,
	0x6b,
	0x64,
	0x63,
	0x65,
	0x6e,
	0x78,
	0x79,
	0x6d,
	0x67,
	0x38,
	0x74,
	0x65,
	0x6e,
	0x72,
	0x67,
	0x68,
	0x6d,
	0x65,
	0x6b,
	0x34,
	0x6e,
	0x6d,
	0x71,
	0x6d,
	0x32,
	0x38,
	0x6b,
	0x2,
	0x4,
	0x0,
	0x0,
	0x20,
	0x0,
	0x5,
	0x0,
	0x21,
	0x0,
	0x7,
	0x0,
	0x24,
	0x0,
	0x3,
	0x0,
	0x23,
	0x0,
	0x7,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x0,
	0x75,
	0x61,
	0x74,
	0x6f,
	0x6d,
	0x31,
	0x32,
	0x33,
	0x34,
	0x35,
	0x36,
	0x37,
	0x66,
	0x6f,
	0x6f,
	0x37,
	0x36,
	0x35,
	0x34,
	0x33,
	0x32,
	0x31,
}
