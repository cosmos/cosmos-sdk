package types_test

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

type errOnMarshal struct {
	testdata.Dog
}

var _ proto.Message = (*errOnMarshal)(nil)

var errAlways = fmt.Errorf("always erroring")

func (eom *errOnMarshal) XXX_Marshal([]byte, bool) ([]byte, error) {
	return nil, errAlways
}

var eom = &errOnMarshal{}

var sink interface{}

func BenchmarkNewAnyWithCustomTypeURLWithErrorReturned(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		any, err := types.NewAnyWithValue(eom)
		if err == nil {
			b.Fatal("err wasn't returned")
		}
		if any != nil {
			b.Fatalf("Unexpectedly got a non-nil Any value: %v", any)
		}
		sink = any
	}
	if sink == nil {
		b.Fatal("benchmark didn't run")
	}
	sink = (interface{})(nil)
}
