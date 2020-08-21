package types_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/stretchr/testify/require"
)

func TestPackClientState(t *testing.T) {
	clientState := localhosttypes.NewClientState(chainID, height)

	clientAny, err := types.PackClientState(clientState)
	require.NoError(t, err, "pack clientstate should not return error")

	cs, err := types.UnpackClientState(clientAny)
	require.NoError(t, err, "unpack clientstate should not return error")

	require.Equal(t, clientState, cs, "client states are not equal after packing and unpacking")

	_, err = types.PackClientState(nil)
	require.Error(t, err, "did not error after packing nil")
}

func TestPackConsensusState(t *testing.T) {
	consensusState := ibctmtypes.NewConsensusState(time.Now(), commitmenttypes.NewMerkleRoot([]byte("root")), height, []byte("nextvalshash"))

	consensusAny, err := types.PackConsensusState(consensusState)
	require.NoError(t, err, "pack consensusstate should not return error")

	cs, err := types.UnpackConsensusState(consensusAny)
	require.NoError(t, err, "unpack consensusstate should not return error")

	require.Equal(t, consensusState, cs, "consensus states are not equal after packing and unpacking")

	_, err = types.PackConsensusState(nil)
	require.Error(t, err, "did not error after packing nil")
}
