package codec

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

type s1 struct {
	i int
}

type s2 struct {
	i interface{}
}

type ss struct {
	b bool
	i int64
	u uint64
}

type complex struct {
	s1 s1
	s2 s2
	ss ss
}

func newComplex() complex {
	return complex{
		s1: s1{rand.Int()},
		s2: s2{s1{rand.Int()}},
		ss: ss{
			b: false,
			i: rand.Int63(),
			u: rand.Uint64(),
		},
	}
}

type tc struct {
	value interface{}
	ptr   func() interface{}
}

func randgen(size int) (res []tc) {
	res = make([]tc, size)
	for i := range res {
		switch rand.Int() % 5 {
		case 0:
			res[i] = tc{rand.Int63(), func() interface{} { return new(int64) }}
		case 1:
			bz := make([]byte, 32)
			rand.Read(bz)
			res[i] = tc{bz, func() interface{} { return new([]byte) }}
		case 2:
			res[i] = tc{&s1{rand.Int()}, func() interface{} { return new(*s1) }}
		case 3:
			res[i] = tc{s2{s1{rand.Int()}}, func() interface{} { return new(s2) }}
		case 4:
			res[i] = tc{newComplex(), func() interface{} { return new(complex) }}
		}
	}
	return
}

func registerCodec(cdc *Amino) {
	cdc.RegisterConcrete(s1{}, "test/s1", nil)
	cdc.RegisterConcrete(s2{}, "test/s2", nil)
}

type marshaller func(interface{}) ([]byte, error)
type unmarshaller func([]byte, interface{}) error
type mustMarshaller func(interface{}) []byte
type mustUnmarshaller func([]byte, interface{})

func testEqual(t *testing.T, tc tc, marshal1 marshaller, unmarshal1 unmarshaller, marshal2 marshaller, unmarshal2 unmarshaller) {
	bz1, err := marshal1(tc.value)
	require.NoError(t, err)
	ptr1 := tc.ptr()
	err = unmarshal1(bz1, ptr1)
	require.NoError(t, err)

	bz2, err := marshal2(tc.value)
	require.NoError(t, err)
	ptr2 := tc.ptr()
	err = unmarshal2(bz2, ptr2)
	require.NoError(t, err)

	require.Equal(t, bz1, bz2)
	require.Equal(t, ptr1, ptr2)
}

func testEqualMust(t *testing.T, tc tc, marshal1 mustMarshaller, unmarshal1 mustUnmarshaller, marshal2 mustMarshaller, unmarshal2 mustUnmarshaller) {
	bz1 := marshal1(tc.value)
	ptr1 := tc.ptr()
	unmarshal1(bz1, ptr1)

	bz2 := marshal2(tc.value)
	ptr2 := tc.ptr()
	unmarshal2(bz2, ptr2)

	require.Equal(t, bz1, bz2)
	require.Equal(t, ptr1, ptr2)
}

func benchmarkAmino(b *testing.B, cdc *Amino, datanum int, lambda int) {
	tcs := randgen(datanum)
	b.ResetTimer()

	exec := executor()
	for i := 0; i < b.N; i++ {
		index := int(rand.ExpFloat64()*float64(datanum)/float64(lambda)) % datanum
		tc := tcs[index]
		ptr := tc.ptr()

		exec(func() {
			bz, _ := cdc.MarshalJSON(tc.value)
			cdc.UnmarshalJSON(bz, ptr)
		})
		exec(func() {
			bz, _ := cdc.MarshalBinary(tc.value)
			cdc.UnmarshalBinary(bz, ptr)
		})
	}

}

const gonum = 5

func executor() func(func()) {
	ch := make(chan bool, gonum)
	return func(f func()) {
		ch <- true
		go func() {
			defer func() { <-ch }()
			f()
		}()
	}
}
