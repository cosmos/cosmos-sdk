package types_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestConfig_SetCoinType(t *testing.T) {
	config := &sdk.Config{}
	require.Equal(t, uint32(0), config.GetCoinType())
	config.SetCoinType(99)
	require.Equal(t, uint32(99), config.GetCoinType())
}

func TestConfig_SetTxEncoder(t *testing.T) {
	mockErr := errors.New("test")
	config := &sdk.Config{}
	require.Nil(t, config.GetTxEncoder())
	encFunc := sdk.TxEncoder(func(tx sdk.Tx) ([]byte, error) { return nil, nil })
	config.SetTxEncoder(encFunc)
	_, err := config.GetTxEncoder()(sdk.Tx(nil))
	require.Error(t, mockErr, err)
}

func TestConfig_SetFullFundraiserPath(t *testing.T) {
	config := &sdk.Config{}
	require.Equal(t, "", config.GetFullFundraiserPath())
	config.SetFullFundraiserPath("test/path")
	require.Equal(t, "test/path", config.GetFullFundraiserPath())
}

func TestKeyringServiceName(t *testing.T) {
	require.Equal(t, sdk.DefaultKeyringServiceName, sdk.KeyringServiceName())
}
