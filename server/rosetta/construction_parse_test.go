package rosetta

import (
	"testing"
)

func TestLaunchpad_ConstructionParse(t *testing.T) {
	//properties := properties{
	//	Blockchain: "TheBlockchain",
	//	Network:    "TheNetwork",
	//}
	//adapter := newAdapter(nil, cosmos.NewClient("", nil), tendermint.NewClient(""), properties)
	//
	//cases := []struct {
	//	name  string
	//	getTx func() string
	//	resp  *types.ConstructionParseResponse
	//	err   *types.Error
	//}{
	//	// TODO: Add a test for unsigned tx
	//	{"invalid tx",
	//		func() string {
	//			return "invalid"
	//		},
	//		nil,
	//		ErrInvalidTransaction,
	//	},
	//}
	//
	//for _, tt := range cases {
	//	tt := tt
	//	t.Run(tt.name, func(t *testing.T) {
	//		req := &types.ConstructionParseRequest{
	//			Transaction: tt.getTx(),
	//		}
	//		parseResp, parseErr := adapter.ConstructionParse(context.Background(), req)
	//		require.Equal(t, tt.err, parseErr)
	//		require.Equal(t, tt.resp, parseResp)
	//	})
	//}
}
