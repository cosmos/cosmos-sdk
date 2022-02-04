package cli

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func TestParseSubmitSoftwareUpgradeProposal(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	expectedMetadata := []byte{42}
	expectedPlan := &types.Plan{
		Name:   "example",
		Height: 123450000,
	}

	okJSON := testutil.WriteToNewTempFile(t, fmt.Sprintf(`
{
	"plan": {
		"name": "%s",
		"height": %d
	},
	"metadata": "%s",
	"deposit": "1000test"
}
`, expectedPlan.Name, expectedPlan.Height, base64.StdEncoding.EncodeToString(expectedMetadata)))

	badJSON := testutil.WriteToNewTempFile(t, "bad json")

	// nonexistent json
	_, _, _, err := parseSubmitSoftwareUpgradeProposal(cdc, "fileDoesNotExist")
	require.Error(t, err)

	// invalid json
	_, _, _, err = parseSubmitSoftwareUpgradeProposal(cdc, badJSON.Name())
	require.Error(t, err)

	// ok json
	upgradePlan, metadata, deposit, err := parseSubmitSoftwareUpgradeProposal(cdc, okJSON.Name())
	require.NoError(t, err, "unexpected error")
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(1000))), deposit)
	require.Equal(t, expectedMetadata, metadata)
	require.Equal(t, upgradePlan, expectedPlan)

	err = okJSON.Close()
	require.Nil(t, err, "unexpected error")
	err = badJSON.Close()
	require.Nil(t, err, "unexpected error")
}
