package server

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/mock"
)

func TestInit(t *testing.T) {
	defer setupViper(t)()

	logger := log.NewNopLogger()
	cmd := InitCmd(mock.GenInitOptions, logger)
	err := cmd.RunE(nil, nil)
	require.NoError(t, err)
}
