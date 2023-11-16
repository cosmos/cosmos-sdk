package wasm

import (
	"encoding/binary"
	"testing"
	"unsafe"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
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
	inPtr, inLen := m.WriteProto(greet)
	outPtr, outLen := m.Exec(inPtr, inLen)
	greetRes := &test1.GreetResponse{}
	m.ReadProtoOut(outPtr, outLen, greetRes)
	if check {
		checkGreetRes(b, greetRes.Message)
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
		checkGreetRes(b, greetRes.Message)
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
	inPtr := m.WriteZeroPB(bz, n)
	outPtr, _ := m.Exec(inPtr, n)
	out, outLen := m.ReadZeroPBOutPtr(outPtr)
	if check {
		checkGreetRes(b, string(out[4:]))
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

func sampleZeroPbGreetCopy(b testing.TB, g test1.Greet) ([]byte, int32) {
	greet := (*GreetZeroPB)(zeroPbBuf)
	greet.Name.Ptr = int16(sizeOfGreetZeroPB)
	greet.Name.Len = uint16(len(g.Name))
	greet.Value = g.Value
	n := len(g.Name)
	strBuf := unsafe.Slice((*byte)(unsafe.Add(zeroPbBuf, sizeOfGreetZeroPB)), n)
	copy(strBuf, g.Name)
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
		checkGreetRes(b, string(out[4:]))
	}
}

func checkGreetRes(b testing.TB, msg string) {
	require.Equal(b, "Hello, Benchmarker! You entered 51", msg)
}

func BenchmarkProtoWasmMsgSend(b *testing.B) {
	m := LoadWasmModule(b, "../../rust/target/wasm32-unknown-unknown/release/example_module2.wasm")
	protoWasmMsgSendRound(b, m, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		protoWasmMsgSendRound(b, m, false)
	}
}

func protoWasmMsgSendRound(b *testing.B, m WasmModule, check bool) {
	inPtr, inLen := m.WriteProto(exampleMsgSend())
	outPtr, outLen := m.ExecMsgSend(inPtr, inLen)
	greetRes := &test1.GreetResponse{}
	m.ReadProtoOut(outPtr, outLen, greetRes)
	if check {
		checkMsgSendRes(b, greetRes.Message)
	}
	m.Free(inPtr, inLen)
	m.Free(outPtr, outLen)
}

func exampleMsgSend() *bankv1beta1.MsgSend {
	coins := []*basev1beta1.Coin{
		{
			Denom:  "uatom",
			Amount: "1234567",
		},
		{
			Denom:  "foo",
			Amount: "7654321",
		},
	}
	return &bankv1beta1.MsgSend{
		FromAddress: "cosmos1huydeevpz37sd9snkgul6070mstupukw00xkw9",
		ToAddress:   "cosmos1xy4yqngt0nlkdcenxymg8tenrghmek4nmqm28k",
		Amount:      coins,
	}
}

func BenchmarkProtoFFIMsgSend(b *testing.B) {
	m := LoadFFIModule(b, "../../rust/target/release/libexample_module2.dylib")
	protoFFIMsgSendRound(b, m, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		protoFFIMsgSendRound(b, m, false)
	}
}

func protoFFIMsgSendRound(b testing.TB, m FFIModule, check bool) {
	msgSend := exampleMsgSend()
	bz, err := proto.Marshal(msgSend)
	if check {
		require.NoError(b, err)
	}
	out := m.ExecMsgSend(bz)
	greetRes := &test1.GreetResponse{}
	err = proto.Unmarshal(out, greetRes)
	if check {
		require.NoError(b, err)
		checkMsgSendRes(b, greetRes.Message)
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
	inPtr := m.WriteZeroPB(testMsgSendZeroPbBz, int32(len(testMsgSendZeroPbBz)))
	outPtr, _ := m.ExecMsgSend(inPtr, int32(len(testMsgSendZeroPbBz)))
	out, outLen := m.ReadZeroPBOutPtr(outPtr)
	if check {
		checkMsgSendZeroPbRes(b, out)
		require.Equal(b, int32(132), outLen)
	}
	m.Free(inPtr, 0x10000)
	m.Free(outPtr, 0x10000)
}

var zeroPbMsgSendBz []byte

func BenchmarkZeroPBFFIMsgSend(b *testing.B) {
	m := LoadFFIModule(b, "../../rust/target/release/libexample_module.dylib")
	n := len(testMsgSendZeroPbBz)
	zeroPbMsgSendBz = unsafe.Slice((*byte)(zeroPbBuf), 0x10000)
	binary.LittleEndian.PutUint16(zeroPbMsgSendBz[0x10000-2:0x10000], uint16(n))
	copy(zeroPbMsgSendBz, testMsgSendZeroPbBz)
	zeroPBFFIMsgSendRound(b, m, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		zeroPBFFIMsgSendRound(b, m, false)
	}
}

func zeroPBFFIMsgSendRound(b testing.TB, m FFIModule, check bool) {
	out := m.ExecMsgSend(zeroPbMsgSendBz)
	if check {
		checkMsgSendZeroPbRes(b, out)
	}
}

func checkMsgSendRes(b testing.TB, msg string) {
	require.Equal(b, "cosmos1huydeevpz37sd9snkgul6070mstupukw00xkw9 sent to cosmos1xy4yqngt0nlkdcenxymg8tenrghmek4nmqm28k:  1234567 uatom  7654321 foo", msg)
}

func checkMsgSendZeroPbRes(b testing.TB, out []byte) {
	checkMsgSendRes(b, string(out[4:]))
	require.Equal(b, []byte{0x77, 0, 0xd, 0}, out[:4])
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

func BenchmarkZeroPBFFIMarshal(b *testing.B) {
	m := LoadFFIModule(b, "../../rust/target/release/libexample_module.dylib")
	zeroPBFFIRound(b, m, true)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		zeroPBFFIMarshalRound(b, m, false)
	}

	b.StopTimer()
	allocations := m.Allocations()
	require.Equal(b, int32(1), allocations)
}

func zeroPBFFIMarshalRound(b testing.TB, m FFIModule, check bool) {
	g := test1.Greet{
		Value: 51,
		Name:  "Benchmarker",
	}
	bz, n := sampleZeroPbGreetCopy(b, g)
	binary.LittleEndian.PutUint16(bz[0x10000-2:0x10000], uint16(n))
	out := m.Exec(bz)
	greetRes := test1.GreetResponse{
		Message: string(out[4:]),
	}
	if check && greetRes.Message != "Hello, Benchmarker! You entered 51" {
		b.Fatal("message mismatch")
	}
	if check {
		require.Equal(b, []byte{4, 0, 34, 0}, out[:4])
		require.Equal(b, "Hello, Benchmarker! You entered 51", greetRes.Message)
	}
}
