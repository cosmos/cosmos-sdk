package types_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *SoloMachineTestSuite) TestCheckHeaderAndUpdateState() {
	var (
		clientState exported.ClientState
		header      exported.Header
	)

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		testCases := []struct {
			name    string
			setup   func()
			expPass bool
		}{
			{
				"successful update",
				func() {
					clientState = solomachine.ClientState()
					header = solomachine.CreateHeader()
				},
				true,
			},
			{
				"wrong client state type",
				func() {
					clientState = &ibctmtypes.ClientState{}
					header = solomachine.CreateHeader()
				},
				false,
			},
			{
				"invalid header type",
				func() {
					clientState = solomachine.ClientState()
					header = &ibctmtypes.Header{}
				},
				false,
			},
			{
				"wrong sequence in header",
				func() {
					clientState = solomachine.ClientState()
					// store in temp before assigning to interface type
					h := solomachine.CreateHeader()
					h.Sequence++
					header = h
				},
				false,
			},
			{
				"invalid header Signature",
				func() {
					clientState = solomachine.ClientState()
					h := solomachine.CreateHeader()
					h.Signature = suite.GetInvalidProof()
					header = h
				}, false,
			},
			{
				"invalid timestamp in header",
				func() {
					clientState = solomachine.ClientState()
					h := solomachine.CreateHeader()
					h.Timestamp--
					header = h
				}, false,
			},
			{
				"signature uses wrong sequence",
				func() {
					clientState = solomachine.ClientState()
					solomachine.Sequence++
					header = solomachine.CreateHeader()
				},
				false,
			},
			{
				"signature uses new pubkey to sign",
				func() {
					// store in temp before assinging to interface type
					cs := solomachine.ClientState()
					h := solomachine.CreateHeader()

					publicKey, err := codectypes.NewAnyWithValue(solomachine.PublicKey)
					suite.NoError(err)

					data := &types.HeaderData{
						NewPubKey:      publicKey,
						NewDiversifier: h.NewDiversifier,
					}

					dataBz, err := suite.chainA.Codec.MarshalBinaryBare(data)
					suite.Require().NoError(err)

					// generate invalid signature
					signBytes := &types.SignBytes{
						Sequence:    cs.Sequence,
						Timestamp:   solomachine.Time,
						Diversifier: solomachine.Diversifier,
						DataType:    types.CLIENT,
						Data:        dataBz,
					}

					signBz, err := suite.chainA.Codec.MarshalBinaryBare(signBytes)
					suite.Require().NoError(err)

					sig := solomachine.GenerateSignature(signBz)
					suite.Require().NoError(err)
					h.Signature = sig

					clientState = cs
					header = h

				},
				false,
			},
			{
				"signature signs over old pubkey",
				func() {
					// store in temp before assinging to interface type
					cs := solomachine.ClientState()
					oldPubKey := solomachine.PublicKey
					h := solomachine.CreateHeader()

					// generate invalid signature
					data := append(sdk.Uint64ToBigEndian(cs.Sequence), oldPubKey.Bytes()...)
					sig := solomachine.GenerateSignature(data)
					h.Signature = sig

					clientState = cs
					header = h
				},
				false,
			},
			{
				"consensus state public key is nil",
				func() {
					cs := solomachine.ClientState()
					cs.ConsensusState.PublicKey = nil
					clientState = cs
					header = solomachine.CreateHeader()
				},
				false,
			},
		}

		for _, tc := range testCases {
			tc := tc

			suite.Run(tc.name, func() {
				// setup test
				tc.setup()

				clientState, consensusState, err := clientState.CheckHeaderAndUpdateState(suite.chainA.GetContext(), suite.chainA.Codec, suite.store, header)

				if tc.expPass {
					suite.Require().NoError(err)
					suite.Require().Equal(header.(*types.Header).NewPublicKey, clientState.(*types.ClientState).ConsensusState.PublicKey)
					suite.Require().Equal(uint64(0), clientState.(*types.ClientState).FrozenSequence)
					suite.Require().Equal(header.(*types.Header).Sequence+1, clientState.(*types.ClientState).Sequence)
					suite.Require().Equal(consensusState, clientState.(*types.ClientState).ConsensusState)
				} else {
					suite.Require().Error(err)
					suite.Require().Nil(clientState)
					suite.Require().Nil(consensusState)
				}
			})
		}
	}
}
