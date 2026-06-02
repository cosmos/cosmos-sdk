package v6

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type mockParamsKeeper struct {
	params types.Params
	err    error
}

func (m *mockParamsKeeper) GetParams(context.Context) (types.Params, error) {
	return m.params, m.err
}

func (m *mockParamsKeeper) SetParams(_ context.Context, params types.Params) error {
	m.params = params
	return nil
}

func TestMigrate(t *testing.T) {
	ctx := context.Background()

	t.Run("sets default fee amount with existing bond denom", func(t *testing.T) {
		params := types.DefaultParams()
		params.BondDenom = "uatom"
		params.KeyRotationFee = sdk.Coin{}
		k := &mockParamsKeeper{params: params}

		require.NoError(t, Migrate(ctx, k))
		require.Equal(t, sdk.NewCoin("uatom", types.DefaultKeyRotationFee.Amount), k.params.KeyRotationFee)
		require.NoError(t, k.params.Validate())
	})

	t.Run("returns invalid existing params error", func(t *testing.T) {
		k := &mockParamsKeeper{}

		err := Migrate(ctx, k)
		require.Error(t, err)
	})
}
