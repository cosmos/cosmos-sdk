package types_test

import (
	"time"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

func (suite *TendermintTestSuite) TestEvidence() {
	signers := []tmtypes.PrivValidator{suite.privVal}

	ev := &types.Evidence{
		Header1:  suite.header,
		Header2:  types.CreateTestHeader(chainID, height, height-1, suite.now, suite.valSet, suite.valSet, signers),
		ChainId:  chainID,
		ClientId: clientID,
	}

	suite.Require().Equal(ev.ClientType(), clientexported.Tendermint)
	suite.Require().Equal(ev.GetClientID(), clientID)
	suite.Require().Equal(ev.Route(), "client")
	suite.Require().Equal(ev.Type(), "client_misbehaviour")
	suite.Require().Equal(ev.Hash(), tmbytes.HexBytes(tmhash.Sum(suite.cdc.MustMarshalBinaryBare(ev))))
	suite.Require().Equal(ev.GetHeight(), int64(height))
}

func (suite *TendermintTestSuite) TestEvidenceValidateBasic() {
	altPrivVal := tmtypes.NewMockPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)

	altVal := tmtypes.NewValidator(altPubKey, height)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	signers := []tmtypes.PrivValidator{suite.privVal}

	// Create signer array and ensure it is in same order as bothValSet
	_, suiteVal := suite.valSet.GetByIndex(0)
	bothSigners := types.CreateSortedSignerArray(altPrivVal, suite.privVal, altVal, suiteVal)

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	testCases := []struct {
		name             string
		evidence         *types.Evidence
		malleateEvidence func(ev *types.Evidence) error
		expPass          bool
	}{
		{
			"valid evidence",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader(chainID, height, height-1, suite.now.Add(time.Minute), suite.valSet, suite.valSet, signers),
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			true,
		},
		{
			"evidence Header1 is nil",
			types.NewEvidence(clientID, chainID, nil, suite.header),
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"evidence Header2 is nil",
			types.NewEvidence(clientID, chainID, suite.header, nil),
			func(ev *types.Evidence) error { return nil },
			false,
		},

		{
			"valid evidence with different trusted headers",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader(chainID, height, height-3, suite.now.Add(time.Minute), suite.valSet, bothValSet, signers),
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			true,
		},
		{
			"trusted height is 0 in Header1",
			&types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, 0, suite.now.Add(time.Minute), suite.valSet, suite.valSet, signers),
				Header2:  suite.header,
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"trusted height is 0 in Header2",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader(chainID, height, 0, suite.now.Add(time.Minute), suite.valSet, suite.valSet, signers),
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"trusted valset is nil in Header1",
			&types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height-1, suite.now.Add(time.Minute), suite.valSet, nil, signers),
				Header2:  suite.header,
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"trusted valset is nil in Header2",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader(chainID, height, height-1, suite.now.Add(time.Minute), suite.valSet, nil, signers),
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"invalid client ID ",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader(chainID, height, height-1, suite.now, suite.valSet, suite.valSet, signers),
				ChainId:  chainID,
				ClientId: "GAIA",
			},
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"wrong chainID on header1",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader("ethermint", height, height-1, suite.now, suite.valSet, suite.valSet, signers),
				ChainId:  "ethermint",
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"wrong chainID on header2",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader("ethermint", height, height-1, suite.now, suite.valSet, suite.valSet, signers),
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"mismatched heights",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader(chainID, 6, 4, suite.now, suite.valSet, suite.valSet, signers),
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"same block id",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  suite.header,
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error { return nil },
			false,
		},
		{
			"header 1 doesn't have 2/3 majority",
			&types.Evidence{
				Header1:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, suite.valSet, bothSigners),
				Header2:  suite.header,
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error {
				// voteSet contains only altVal which is less than 2/3 of total power (height/1height)
				wrongVoteSet := tmtypes.NewVoteSet(chainID, int64(ev.Header1.GetHeight()), 1, tmproto.PrecommitType, altValSet)
				blockID, err := tmtypes.BlockIDFromProto(&ev.Header1.Commit.BlockID)
				if err != nil {
					return err
				}

				tmCommit, err := tmtypes.MakeCommit(*blockID, int64(ev.Header2.GetHeight()), ev.Header1.Commit.Round, wrongVoteSet, altSigners, suite.now)
				ev.Header1.Commit = tmCommit.ToProto()
				return err
			},
			false,
		},
		{
			"header 2 doesn't have 2/3 majority",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, suite.valSet, bothSigners),
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error {
				// voteSet contains only altVal which is less than 2/3 of total power (height/1height)
				wrongVoteSet := tmtypes.NewVoteSet(chainID, int64(ev.Header2.GetHeight()), 1, tmproto.PrecommitType, altValSet)
				blockID, err := tmtypes.BlockIDFromProto(&ev.Header2.Commit.BlockID)
				if err != nil {
					return err
				}

				tmCommit, err := tmtypes.MakeCommit(*blockID, int64(ev.Header2.GetHeight()), ev.Header2.Commit.Round, wrongVoteSet, altSigners, suite.now)
				ev.Header2.Commit = tmCommit.ToProto()
				return err
			},
			false,
		},
		{
			"validators sign off on wrong commit",
			&types.Evidence{
				Header1:  suite.header,
				Header2:  types.CreateTestHeader(chainID, height, height-1, suite.now, bothValSet, suite.valSet, bothSigners),
				ChainId:  chainID,
				ClientId: clientID,
			},
			func(ev *types.Evidence) error {
				tmBlockID := types.MakeBlockID(tmhash.Sum([]byte("other_hash")), 3, tmhash.Sum([]byte("other_partset")))
				ev.Header2.Commit.BlockID = tmBlockID.ToProto()
				return nil
			},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.malleateEvidence(tc.evidence)
		suite.Require().NoError(err)

		if tc.expPass {
			suite.Require().NoError(tc.evidence.ValidateBasic(), "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(tc.evidence.ValidateBasic(), "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
