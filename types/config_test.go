package types_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestConfig_SetCoinType(t *testing.T) {
	config := sdk.NewConfig()
	config.SetCoinType(1)
	require.Equal(t, uint32(1), config.GetCoinType())
	config.SetCoinType(99)
	require.Equal(t, uint32(99), config.GetCoinType())

	config.Seal()
	require.Panics(t, func() { config.SetCoinType(99) })
}

func TestConfig_SetTxEncoder(t *testing.T) {
	mockErr := errors.New("test")
	config := sdk.NewConfig()
	require.Nil(t, config.GetTxEncoder())
	encFunc := sdk.TxEncoder(func(tx sdk.Tx) ([]byte, error) { return nil, nil })
	config.SetTxEncoder(encFunc)
	_, err := config.GetTxEncoder()(sdk.Tx(nil))
	require.Error(t, mockErr, err)

	config.Seal()
	require.Panics(t, func() { config.SetTxEncoder(encFunc) })
}

func TestConfig_SetFullFundraiserPath(t *testing.T) {
	config := sdk.NewConfig()
	config.SetFullFundraiserPath("test/path")
	require.Equal(t, "test/path", config.GetFullFundraiserPath())

	config.SetFullFundraiserPath("test/poth")
	require.Equal(t, "test/poth", config.GetFullFundraiserPath())

	config.Seal()
	require.Panics(t, func() { config.SetFullFundraiserPath("x/test/path") })
}

func TestKeyringServiceName(t *testing.T) {
	require.Equal(t, sdk.DefaultKeyringServiceName, sdk.KeyringServiceName())
}
