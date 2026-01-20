package types_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

type errOnMarshal struct {
	testdata.Dog
}

var _ proto.Message = (*errOnMarshal)(nil)

var errAlways = fmt.Errorf("always erroring")

func (eom *errOnMarshal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return nil, errAlways
}

var eom = &errOnMarshal{}

// Ensure that returning an error doesn't suddenly allocate and waste bytes.
// See https://github.com/cosmos/cosmos-sdk/issues/8537
func TestNewAnyWithCustomTypeURLWithErrorNoAllocation(t *testing.T) {
	// This test continues to fail inconsistently.
	//
	// Example: https://github.com/cosmos/cosmos-sdk/pull/9246/checks?check_run_id=2643313958#step:6:118
	// Ref: https://github.com/cosmos/cosmos-sdk/issues/9010
	t.SkipNow()

	// make sure we're not in the middle of a GC.
	runtime.GC()

	var ms1, ms2 runtime.MemStats
	runtime.ReadMemStats(&ms1)
	any, err := types.NewAnyWithValue(eom)
	runtime.ReadMemStats(&ms2)
	// Ensure that no fresh allocation was made.
	if diff := ms2.HeapAlloc - ms1.HeapAlloc; diff > 0 {
		t.Errorf("Unexpected allocation of %d bytes", diff)
	}
	if err == nil {
		t.Fatal("err wasn't returned")
	}
	if any != nil {
		t.Fatalf("Unexpectedly got a non-nil Any value: %v", any)
	}
}

var sink any

func BenchmarkNewAnyWithCustomTypeURLWithErrorReturned(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
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
	sink = (any)(nil)
}

// TestUnpackAnyWithEmptyValue tests the fix for handling empty any.Value.
// Before the fix, when any.Value was nil or empty, proto.Unmarshal would succeed
// (because unmarshaling empty data is valid - it creates an empty message),
// but then we'd be assigning a nil/empty value which could cause issues.
// The fix adds a check for empty any.Value before attempting to unmarshal.
func TestUnpackAnyWithEmptyValue(t *testing.T) {
	registry := types.NewInterfaceRegistry()
	registry.RegisterInterface("Animal", (*testdata.Animal)(nil))
	registry.RegisterImplementations(
		(*testdata.Animal)(nil),
		&testdata.Dog{},
		&testdata.Cat{},
	)

	t.Run("nil Value returns nil without error", func(t *testing.T) {
		any := &types.Any{
			TypeUrl: "/testdata.Dog",
			Value:   nil,
		}

		var animal testdata.Animal
		err := registry.UnpackAny(any, &animal)
		require.NoError(t, err)
		require.Nil(t, animal, "animal should be nil when Value is nil")
	})

	t.Run("empty slice Value returns nil without error", func(t *testing.T) {
		any := &types.Any{
			TypeUrl: "/testdata.Dog",
			Value:   []byte{},
		}

		var animal testdata.Animal
		err := registry.UnpackAny(any, &animal)
		require.NoError(t, err)
		require.Nil(t, animal, "animal should be nil when Value is empty")
	})

	t.Run("valid Value unmarshals correctly", func(t *testing.T) {
		// Create a valid Any with actual data
		dog := &testdata.Dog{Name: "Rufus"}
		any, err := types.NewAnyWithValue(dog)
		require.NoError(t, err)
		require.NotNil(t, any)
		require.NotEmpty(t, any.Value, "Value should not be empty for valid message")

		var animal testdata.Animal
		err = registry.UnpackAny(any, &animal)
		require.NoError(t, err)
		require.NotNil(t, animal, "animal should not be nil when Value has valid data")
		require.Equal(t, dog, animal)
	})

	t.Run("empty Value with empty TypeUrl returns nil", func(t *testing.T) {
		any := &types.Any{
			TypeUrl: "",
			Value:   []byte{},
		}

		var animal testdata.Animal
		err := registry.UnpackAny(any, &animal)
		require.NoError(t, err)
		require.Nil(t, animal, "animal should be nil when TypeUrl is empty")
	})

	t.Run("nil Value with empty TypeUrl returns nil", func(t *testing.T) {
		any := &types.Any{
			TypeUrl: "",
			Value:   nil,
		}

		var animal testdata.Animal
		err := registry.UnpackAny(any, &animal)
		require.NoError(t, err)
		require.Nil(t, animal, "animal should be nil when TypeUrl is empty")
	})
}
