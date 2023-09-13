package stablejson_test

import (
	"testing"

	"github.com/cosmos/cosmos-proto/anyutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/orm/internal/stablejson"
)

func TestStableJSON(t *testing.T) {
	msg, err := anyutil.New(&bankv1beta1.MsgSend{
		FromAddress: "foo213325",
		ToAddress:   "foo32t5sdfh",
		Amount: []*basev1beta1.Coin{
			{
				Denom:  "bar",
				Amount: "1234",
			},
			{
				Denom:  "baz",
				Amount: "321",
			},
		},
	})
	require.NoError(t, err)
	bz, err := stablejson.Marshal(&txv1beta1.TxBody{Messages: []*anypb.Any{msg}})
	require.NoError(t, err)
	require.Equal(t,
		`{"messages":[{"@type":"cosmos.bank.v1beta1.MsgSend","from_address":"foo213325","to_address":"foo32t5sdfh","amount":[{"denom":"bar","amount":"1234"},{"denom":"baz","amount":"321"}]}]}`,
		string(bz))
}
