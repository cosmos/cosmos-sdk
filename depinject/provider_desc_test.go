package depinject

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type StructIn struct {
	In
	X int
	Y float64 `optional:"true"`
}

type BadOptional struct {
	In
	X int `optional:"foo"`
}

type StructOut struct {
	Out
	X string
	Y []byte
}

func TestExtractProviderDescriptor(t *testing.T) {
	var (
		intType     = reflect.TypeOf(0)
		int16Type   = reflect.TypeOf(int16(0))
		int32Type   = reflect.TypeOf(int32(0))
		float32Type = reflect.TypeOf(float32(0.0))
		float64Type = reflect.TypeOf(0.0)
		stringType  = reflect.TypeOf("")
		byteTyp     = reflect.TypeOf(byte(0))
		bytesTyp    = reflect.TypeOf([]byte{})
	)

	tests := []struct {
		name    string
		ctr     interface{}
		wantIn  []providerInput
		wantOut []providerOutput
		wantErr bool
	}{
		{
			"simple args",
			func(x int, y float64) (string, []byte) { return "", nil },
			[]providerInput{{Type: intType}, {Type: float64Type}},
			[]providerOutput{{Type: stringType}, {Type: bytesTyp}},
			false,
		},
		{
			"simple args with error",
			func(x int, y float64) (string, []byte, error) { return "", nil, nil },
			[]providerInput{{Type: intType}, {Type: float64Type}},
			[]providerOutput{{Type: stringType}, {Type: bytesTyp}},
			false,
		},
		{
			"struct in and out",
			func(_ float32, _ StructIn, _ byte) (int16, StructOut, int32, error) {
				return int16(0), StructOut{}, int32(0), nil
			},
			[]providerInput{{Type: float32Type}, {Type: intType}, {Type: float64Type, Optional: true}, {Type: byteTyp}},
			[]providerOutput{{Type: int16Type}, {Type: stringType}, {Type: bytesTyp}, {Type: int32Type}},
			false,
		},
		{
			"error bad position",
			func() (error, int) { return nil, 0 },
			nil,
			nil,
			true,
		},
		{
			"bad optional",
			func(_ BadOptional) int { return 0 },
			nil,
			nil,
			true,
		},
		{
			"variadic",
			func(...float64) int { return 0 },
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractProviderDescriptor(tt.ctr)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractProviderDescriptor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Inputs, tt.wantIn) {
				t.Errorf("extractProviderDescriptor() got = %v, want %v", got.Inputs, tt.wantIn)
			}
			if !reflect.DeepEqual(got.Outputs, tt.wantOut) {
				t.Errorf("extractProviderDescriptor() got = %v, want %v", got.Outputs, tt.wantOut)
			}
		})
	}
}

type SomeStruct struct{}

func TestBadCtr(t *testing.T) {
	_, err := extractProviderDescriptor(SomeStruct{})
	require.Error(t, err)
}

func TestErrorFunc(t *testing.T) {
	_, err := extractProviderDescriptor(
		func() (error, int) { return nil, 0 },
	)
	require.Error(t, err)

	_, err = extractProviderDescriptor(
		func() (int, error) { return 0, nil },
	)
	require.NoError(t, err)

	var x int
	require.Error(t,
		Inject(
			Provide(func() (int, error) {
				return 0, fmt.Errorf("the error")
			}),
			&x,
		))
}
