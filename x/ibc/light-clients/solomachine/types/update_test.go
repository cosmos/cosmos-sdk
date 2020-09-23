package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
)

func (suite *SoloMachineTestSuite) TestCheckHeaderAndUpdateState() {
	var (
		clientState exported.ClientState
		header      exported.Header
	)

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			"successful update",
			func() {
				clientState = suite.solomachine.ClientState()
				header = suite.solomachine.CreateHeader()
			},
			true,
		},
		{
			"wrong client state type",
			func() {
				clientState = &ibctmtypes.ClientState{}
				header = suite.solomachine.CreateHeader()
			},
			false,
		},
		{
			"invalid header type",
			func() {
				clientState = suite.solomachine.ClientState()
				header = &ibctmtypes.Header{}
			},
			false,
		},
		{
			"wrong sequence in header",
			func() {
				clientState = suite.solomachine.ClientState()
				// store in temp before assigning to interface type
				h := suite.solomachine.CreateHeader()
				h.Sequence++
				header = h
			},
			false,
		},
		{
			"signature uses wrong sequence",
			func() {
				clientState = suite.solomachine.ClientState()
				suite.solomachine.Sequence++
				header = suite.solomachine.CreateHeader()
			},
			false,
		},
		{
			"signature uses new pubkey to sign",
			func() {
				// store in temp before assinging to interface type
				cs := suite.solomachine.ClientState()
				h := suite.solomachine.CreateHeader()

				publicKey, err := tx.PubKeyToAny(suite.solomachine.PublicKey)
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
					Timestamp:   suite.solomachine.Time,
					Diversifier: suite.solomachine.Diversifier,
					Data:        dataBz,
				}

				signBz, err := suite.chainA.Codec.MarshalBinaryBare(signBytes)
				suite.Require().NoError(err)

				sig, err := suite.solomachine.PrivateKey.Sign(signBz)
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
				cs := suite.solomachine.ClientState()
				oldPrivKey := suite.solomachine.PrivateKey
				h := suite.solomachine.CreateHeader()

				// generate invalid signature
				data := append(sdk.Uint64ToBigEndian(cs.Sequence), oldPrivKey.PubKey().Bytes()...)
				sig, err := oldPrivKey.Sign(data)
				suite.Require().NoError(err)
				h.Signature = sig

				clientState = cs
				header = h
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
