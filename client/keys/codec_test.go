package keys

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCases struct {
	Keys    []KeyOutput
	Answers []KeyOutput
	JSON    [][]byte
}

func getTestCases() testCases {
	return testCases{
		[]KeyOutput{
			{"A", "B", "C", "D", "E"},
			{"A", "B", "C", "D", ""},
			{"", "B", "C", "D", ""},
			{"", "", "", "", ""},
		},
		make([]KeyOutput, 4),
		[][]byte{
			[]byte(`{"name":"A","type":"B","address":"C","pub_key":"D","mnemonic":"E"}`),
			[]byte(`{"name":"A","type":"B","address":"C","pub_key":"D"}`),
			[]byte(`{"name":"","type":"B","address":"C","pub_key":"D"}`),
			[]byte(`{"name":"","type":"","address":"","pub_key":""}`),
		},
	}
}

func TestMarshalJSON(t *testing.T) {
	type args struct {
		o KeyOutput
	}

	data := getTestCases()

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"basic", args{data.Keys[0]}, []byte(data.JSON[0]), false},
		{"mnemonic is optional", args{data.Keys[1]}, []byte(data.JSON[1]), false},

		// REVIEW: Are the next results expected??
		{"empty name", args{data.Keys[2]}, []byte(data.JSON[2]), false},
		{"empty object", args{data.Keys[3]}, []byte(data.JSON[3]), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalJSON(tt.args.o)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Printf("%s\n", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() = %v, want %v", got, tt.want)
			}
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
		t.Run(tt.name, func(t *testing.T) {
			if err := UnmarshalJSON(tt.args.bz, tt.args.ptr); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Confirm deserialized objects are the same
			require.Equal(t, data.Keys[idx], data.Answers[idx])
		})
	}
}
