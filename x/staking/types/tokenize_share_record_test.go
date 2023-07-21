package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestTokenizeShareRecordGetModuleAddress(t *testing.T) {
	emptyAddress := sdk.AccAddress(make([]byte, 20))
	emptyValAddress := sdk.ValAddress(make([]byte, 20))

	record := TokenizeShareRecord{
		Id:            1,
		Owner:         emptyAddress.String(),
		ModuleAccount: fmt.Sprintf("%s%d", TokenizeShareModuleAccountPrefix, 1),
		Validator:     emptyValAddress.String(),
	}

	require.Equal(t, "cosmos1uk7xy36pvkn3legl7pw0sl29q5jtdg0jfxv58eskx4u80a2ekdksvtqsnl", record.GetModuleAddress().String())
}

func TestTokenizeShareRecordGetShareTokenDenom(t *testing.T) {
	emptyAddress := sdk.AccAddress(make([]byte, 20))
	emptyValAddress := sdk.ValAddress(make([]byte, 20))

	record := TokenizeShareRecord{
		Id:            1,
		Owner:         emptyAddress.String(),
		ModuleAccount: fmt.Sprintf("%s%d", TokenizeShareModuleAccountPrefix, 1),
		Validator:     emptyValAddress.String(),
	}

	require.Equal(t, "cosmosvaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqkh52tw/1", record.GetShareTokenDenom())
}

func TestParseShareTokenDenom(t *testing.T) {
	emptyValAddress := sdk.ValAddress(make([]byte, 20))

	record, err := ParseShareTokenDenom("cosmosvaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqkh52tw/1")
	require.NoError(t, err, "should parse ok")

	require.Equal(t, uint64(1), record.Id)
	require.Equal(t, emptyValAddress.String(), record.Validator)
	require.Equal(t, fmt.Sprintf("%s%d", TokenizeShareModuleAccountPrefix, 1), record.ModuleAccount)

	_, err = ParseShareTokenDenom("ibc/cosmosvaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqkh52tw/1")
	require.ErrorContains(t, err, "wrong number of segments")

	_, err = ParseShareTokenDenom("cosmosvaloper2kek/1")
	require.ErrorContains(t, err, "parse val address part")

	_, err = ParseShareTokenDenom("cosmosvaloper1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqkh52tw/$")
	require.ErrorContains(t, err, "parse recordId part")
}
