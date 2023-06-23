package aminojson

import (
	"fmt"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/x/tx/signing/aminojson"

	"github.com/cosmos/cosmos-sdk/codec"
	gogopb "github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/gogo/testpb"
	pulsarpb "github.com/cosmos/cosmos-sdk/tests/integration/tx/internal/pulsar/testpb"
)

func TestRepeatedFields(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	aj := aminojson.NewEncoder(aminojson.EncoderOptions{DoNotSortFields: true})

	cases := map[string]struct {
		gogo    gogoproto.Message
		pulsar  proto.Message
		unequal bool
		errs    bool
	}{
		"unsupported_empty_sets": {
			gogo:    &gogopb.TestRepeatedFields{},
			pulsar:  &pulsarpb.TestRepeatedFields{},
			unequal: true,
		},
		"unsupported_empty_sets_are_set": {
			gogo: &gogopb.TestRepeatedFields{
				NullableDontOmitempty: []*gogopb.Streng{{Value: "foo"}},
				NonNullableOmitempty:  []gogopb.Streng{{Value: "foo"}},
			},
			pulsar: &pulsarpb.TestRepeatedFields{
				NullableDontOmitempty: []*pulsarpb.Streng{{Value: "foo"}},
				NonNullableOmitempty:  []*pulsarpb.Streng{{Value: "foo"}},
			},
		},
		"unsupported_nullable": {
			gogo:   &gogopb.TestNullableFields{},
			pulsar: &pulsarpb.TestNullableFields{},
			errs:   true,
		},
		"unsupported_nullable_set": {
			gogo: &gogopb.TestNullableFields{
				NullableDontOmitempty:    &gogopb.Streng{Value: "foo"},
				NonNullableDontOmitempty: gogopb.Streng{Value: "foo"},
			},
			pulsar: &pulsarpb.TestNullableFields{
				NullableDontOmitempty:    &pulsarpb.Streng{Value: "foo"},
				NonNullableDontOmitempty: &pulsarpb.Streng{Value: "foo"},
			},
			unequal: true,
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			gogoBz, err := cdc.MarshalJSON(tc.gogo)
			require.NoError(t, err)
			pulsarBz, err := aj.Marshal(tc.pulsar)
			if tc.errs {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			fmt.Printf("  gogo: %s\npulsar: %s\n", string(gogoBz), string(pulsarBz))

			if tc.unequal {
				require.NotEqual(t, string(gogoBz), string(pulsarBz))
			} else {
				require.Equal(t, string(gogoBz), string(pulsarBz))
			}
		})
	}
}
