package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func TestNewUpdateClientProposal(t *testing.T) {
	p, err := types.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, clientID, &ibctmtypes.Header{})
	require.NoError(t, err)
	require.NotNil(t, p)

	p, err = types.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, clientID, nil)
	require.Error(t, err)
	require.Nil(t, p)
}

func TestValidateBasic(t *testing.T) {
	// use solo machine header for testing
	solomachine := ibctesting.NewSolomachine(t, clientID)
	smHeader := solomachine.CreateHeader()
	header, err := types.PackHeader(smHeader)
	require.NoError(t, err)

	// use a different pointer so we don't modify 'header'
	smInvalidHeader := solomachine.CreateHeader()

	// a sequence of 0 will fail basic validation
	smInvalidHeader.Sequence = 0

	invalidHeader, err := types.PackHeader(smInvalidHeader)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		proposal govtypes.Content
		expPass  bool
	}{
		{
			"success",
			&types.ClientUpdateProposal{ibctesting.Title, ibctesting.Description, clientID, header},
			true,
		},
		{
			"fails validate abstract - empty title",
			&types.ClientUpdateProposal{"", ibctesting.Description, clientID, header},
			false,
		},
		{
			"fails to unpack header",
			&types.ClientUpdateProposal{ibctesting.Title, ibctesting.Description, clientID, nil},
			false,
		},
		{
			"fails header validate basic",
			&types.ClientUpdateProposal{ibctesting.Title, ibctesting.Description, clientID, invalidHeader},
			false,
		},
	}

	for _, tc := range testCases {

		err := tc.proposal.ValidateBasic()

		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}

	}
}
