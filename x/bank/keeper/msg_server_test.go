package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

func TestUpdateDenomMetadata(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	msgServer := keeper.NewMsgServerImpl(app.BankKeeper)
	testCases := []struct {
		Name        string
		ExpectErr   bool
		ExpectedErr string
		req         types.MsgUpdateDenomMetadata
	}{
		{
			Name:        "fail - wrong authority",
			ExpectErr:   true,
			ExpectedErr: "expected cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn got nongovaccount: expected gov account as only signer for proposal message",
			req: types.MsgUpdateDenomMetadata{
				FromAddress: "nongovaccount",
				Title:       "title",
				Description: "description",
				Metadata: &types.Metadata{
					Name:        "diamondback",
					Symbol:      "DB",
					Description: "The native staking token",
					DenomUnits: []*types.DenomUnit{
						{"udiamondback", uint32(0), []string{"microdiamondback"}},
					},
				},
			},
		},
		{
			Name:        "success - correct authority",
			ExpectErr:   false,
			ExpectedErr: "",
			req: types.MsgUpdateDenomMetadata{
				FromAddress: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
				Title:       "title",
				Description: "description",
				Metadata: &types.Metadata{
					Base:        "diamondback",
					Name:        "diamondback",
					Symbol:      "DB",
					Description: "The native staking token",
					DenomUnits: []*types.DenomUnit{
						{"udiamondback", uint32(0), []string{"microdiamondback"}},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := msgServer.UpdateDenomMetadata(ctx, &testCase.req)
			if testCase.ExpectErr {
				require.Error(t, err)
				require.Equal(t, testCase.ExpectedErr, err.Error())
			} else {
				require.NoError(t, err)
				metadata, _ := app.BankKeeper.GetDenomMetaData(ctx, "diamondback")
				require.Equal(t, testCase.req.Metadata.String(), metadata.String())
			}
		})
	}
}
