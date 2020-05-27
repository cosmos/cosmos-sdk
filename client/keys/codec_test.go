package keys_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

type testCases struct {
	Keys    []keyring.KeyOutput
	Answers []keyring.KeyOutput
	JSON    [][]byte
}

func getTestCases() testCases {
	return testCases{
		// nolint:govet
		[]keyring.KeyOutput{
			{"A", "B", "C", "D", "E", 0, nil},
			{"A", "B", "C", "D", "", 0, nil},
			{"", "B", "C", "D", "", 0, nil},
			{"", "", "", "", "", 0, nil},
		},
		make([]keyring.KeyOutput, 4),
		[][]byte{
			[]byte(`{"name":"A","type":"B","address":"C","pubkey":"D","mnemonic":"E"}`),
			[]byte(`{"name":"A","type":"B","address":"C","pubkey":"D"}`),
			[]byte(`{"name":"","type":"B","address":"C","pubkey":"D"}`),
			[]byte(`{"name":"","type":"","address":"","pubkey":""}`),
		},
	}
}

func TestMarshalJSON(t *testing.T) {
	type args struct {
		o keyring.KeyOutput
	}

	data := getTestCases()

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"basic", args{data.Keys[0]}, data.JSON[0], false},
		{"mnemonic is optional", args{data.Keys[1]}, data.JSON[1], false},

		// REVIEW: Are the next results expected??
		{"empty name", args{data.Keys[2]}, data.JSON[2], false},
		{"empty object", args{data.Keys[3]}, data.JSON[3], false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := keys.MarshalJSON(tt.args.o)
			require.Equal(t, tt.wantErr, err != nil)
			require.True(t, bytes.Equal(got, tt.want))
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	type args struct {
		bz  []byte
		ptr interface{}
	}

	data := getTestCases()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"basic", args{data.JSON[0], &data.Answers[0]}, false},
		{"mnemonic is optional", args{data.JSON[1], &data.Answers[1]}, false},

		// REVIEW: Are the next results expected??
		{"empty name", args{data.JSON[2], &data.Answers[2]}, false},
		{"empty object", args{data.JSON[3], &data.Answers[3]}, false},
	}
	for idx, tt := range tests {
		idx, tt := idx, tt
		t.Run(tt.name, func(t *testing.T) {
			err := keys.UnmarshalJSON(tt.args.bz, tt.args.ptr)
			require.Equal(t, tt.wantErr, err != nil)
			// Confirm deserialized objects are the same
			require.Equal(t, data.Keys[idx], data.Answers[idx])
		})
	}
}
