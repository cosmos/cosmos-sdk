package container_test

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/container"
)

type StructIn struct {
	container.In
	X int
	Y float64 `optional:"true"`
}

type BadOptional struct {
	container.In
	X int `optional:"foo"`
}

type StructOut struct {
	container.Out
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
		wantIn  []container.ProviderInput
		wantOut []container.ProviderOutput
		wantErr bool
	}{
		{
			"simple args",
			func(x int, y float64) (string, []byte) { return "", nil },
			[]container.ProviderInput{{Type: intType}, {Type: float64Type}},
			[]container.ProviderOutput{{Type: stringType}, {Type: bytesTyp}},
			false,
		},
		{
			"simple args with error",
			func(x int, y float64) (string, []byte, error) { return "", nil, nil },
			[]container.ProviderInput{{Type: intType}, {Type: float64Type}},
			[]container.ProviderOutput{{Type: stringType}, {Type: bytesTyp}},
			false,
		},
		{
			"struct in and out",
			func(_ float32, _ StructIn, _ byte) (int16, StructOut, int32, error) {
				return int16(0), StructOut{}, int32(0), nil
			},
			[]container.ProviderInput{{Type: float32Type}, {Type: intType}, {Type: float64Type, Optional: true}, {Type: byteTyp}},
			[]container.ProviderOutput{{Type: int16Type}, {Type: stringType}, {Type: bytesTyp}, {Type: int32Type}},
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
			got, err := container.ExtractProviderDescriptor(tt.ctr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractProviderDescriptor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Inputs, tt.wantIn) {
				t.Errorf("ExtractProviderDescriptor() got = %v, want %v", got.Inputs, tt.wantIn)
			}
			if !reflect.DeepEqual(got.Outputs, tt.wantOut) {
				t.Errorf("ExtractProviderDescriptor() got = %v, want %v", got.Outputs, tt.wantOut)
			}
		})
	}
}
