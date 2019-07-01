package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/require"
)

func TestSupplyMarshalYAML(t *testing.T) {
	supply := DefaultSupply()
	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt()))
	supply.Inflate(coins)

	bz, err := yaml.Marshal(supply)
	require.NoError(t, err)
	bzCoins, err := yaml.Marshal(coins)
	require.NoError(t, err)

	want := fmt.Sprintf(`total:
%s`, string(bzCoins))

	require.Equal(t, want, string(bz))
	require.Equal(t, want, supply.String())
}
