package distribution_test

import (
	"testing"

	simapp "github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// test msg registration
func TestWithdrawTokenizeShareRecordReward(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	h := distribution.NewHandler(app.DistrKeeper)
	delAddr1 = sdk.AccAddress(delPk1.Address())

	res, err := h(ctx, &types.MsgWithdrawAllTokenizeShareRecordReward{
		OwnerAddress: delAddr1.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
}
