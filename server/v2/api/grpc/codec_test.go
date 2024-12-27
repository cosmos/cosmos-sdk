package grpc

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type mockInterfaceRegistry struct{}

func (m mockInterfaceRegistry) Resolve(typeUrl string) (interface {
	ProtoMessage()
	Reset()
	String() string
}, error) {
	return nil, nil
}

func (m mockInterfaceRegistry) RegisterInterface(string, interface{}, ...interface{}) {}
func (m mockInterfaceRegistry) RegisterImplementations(interface{}, ...interface{})   {}
func (m mockInterfaceRegistry) ListAllInterfaces() []string                           { return nil }
func (m mockInterfaceRegistry) ListImplementations(string) []string                   { return nil }
func (m mockInterfaceRegistry) UnpackAny(*anypb.Any, interface{}) error               { return nil }

func TestProtoCodec_Marshal(t *testing.T) {
	registry := mockInterfaceRegistry{}
	codec := newProtoCodec(registry)

	tests := []struct {
		name    string
		input   proto.Message
		wantErr bool
	}{
		{
			name:    "nil message",
			input:   nil,
			wantErr: false,
		},
		{
			name:    "empty message",
			input:   &anypb.Any{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz, err := codec.Marshal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.input == nil {
				require.Empty(t, bz)
			}
		})
	}
}

func TestGRPCCodec_Marshal(t *testing.T) {
	registry := mockInterfaceRegistry{}
	pc := newProtoCodec(registry)
	codec := pc.GRPCCodec()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "protov2 message",
			input:   &anypb.Any{},
			wantErr: false,
		},
		{
			name:    "invalid type",
			input:   "not a proto message",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz, err := codec.Marshal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, errUnknownProtoType)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, bz)
		})
	}
}

func TestGRPCCodec_Unmarshal(t *testing.T) {
	registry := mockInterfaceRegistry{}
	pc := newProtoCodec(registry)
	codec := pc.GRPCCodec()

	msg := &anypb.Any{}
	bz, err := protov2.Marshal(msg)
	require.NoError(t, err)

	tests := []struct {
		name    string
		data    []byte
		msg     interface{}
		wantErr bool
	}{
		{
			name:    "protov2 message",
			data:    bz,
			msg:     &anypb.Any{},
			wantErr: false,
		},
		{
			name:    "invalid type",
			data:    bz,
			msg:     &struct{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := codec.Unmarshal(tt.data, tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, errUnknownProtoType)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCodecName(t *testing.T) {
	registry := mockInterfaceRegistry{}
	pc := newProtoCodec(registry)
	require.Equal(t, "cosmos-sdk-grpc-codec", pc.Name())

	grpcCodec := pc.GRPCCodec()
	require.Equal(t, "cosmos-sdk-grpc-codec", grpcCodec.Name())
}
