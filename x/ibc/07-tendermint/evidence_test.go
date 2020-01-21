package tendermint_test

import (
	"bytes"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
)

func (suite *TendermintTestSuite) TestEvidence() {
	signers := []tmtypes.PrivValidator{suite.privVal}

	ev := tendermint.Evidence{
		Header1:          suite.header,
		Header2:          tendermint.CreateTestHeader(chainID, height, suite.valSet, suite.valSet, signers),
		FromValidatorSet: suite.valSet,
		ChainID:          chainID,
		ClientID:         "gaiamainnet",
	}

	suite.Require().Equal(ev.ClientType(), clientexported.Tendermint)
	suite.Require().Equal(ev.GetClientID(), "gaiamainnet")
	suite.Require().Equal(ev.Type(), "client_misbehaviour")
	// suite.Require().Equal(ev.Hash(), tmhash.Sum(tendermint.SubModuleCdc.MustMarshalBinaryBare(ev))) // FIXME:

}

func (suite *TendermintTestSuite) TestEvidenceValidateBasic() {
	altPrivVal := tmtypes.NewMockPV()
	altVal := tmtypes.NewValidator(altPrivVal.GetPubKey(), height)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	signers := []tmtypes.PrivValidator{suite.privVal}
	// Create signer array and ensure it is in same order as bothValSet
	var bothSigners []tmtypes.PrivValidator
	if bytes.Compare(altPrivVal.GetPubKey().Address(), suite.privVal.GetPubKey().Address()) == -1 {
		bothSigners = []tmtypes.PrivValidator{altPrivVal, suite.privVal}
	} else {
		bothSigners = []tmtypes.PrivValidator{suite.privVal, altPrivVal}
	}

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	testCases := []struct {
		name             string
		evidence         tendermint.Evidence
		malleateEvidence func(ev *tendermint.Evidence) error
		expPass          bool
	}{
		{
			"valid evidence",
			tendermint.Evidence{
				Header1:          suite.header,
				Header2:          tendermint.CreateTestHeader(chainID, height, suite.valSet, bothValSet, signers),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         "gaiamainnet",
			},
			func(ev *tendermint.Evidence) error { return nil },
			true,
		},
		{
			"invalid client ID ",
			tendermint.Evidence{
				Header1:          suite.header,
				Header2:          tendermint.CreateTestHeader(chainID, height, suite.valSet, bothValSet, signers),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         "GAIA",
			},
			func(ev *tendermint.Evidence) error { return nil },
			false,
		},
		{
			"wrong chainID on header1",
			tendermint.Evidence{
				Header1:          suite.header,
				Header2:          tendermint.CreateTestHeader("ethermint", height, suite.valSet, bothValSet, signers),
				FromValidatorSet: bothValSet,
				ChainID:          "ethermint",
				ClientID:         "gaiamainnet",
			},
			func(ev *tendermint.Evidence) error { return nil },
			false,
		},
		{
			"wrong chainID on header2",
			tendermint.Evidence{
				Header1:          suite.header,
				Header2:          tendermint.CreateTestHeader("ethermint", height, suite.valSet, bothValSet, signers),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         "gaiamainnet",
			},
			func(ev *tendermint.Evidence) error { return nil },
			false,
		},
		{
			"mismatched heights",
			tendermint.Evidence{
				Header1:          suite.header,
				Header2:          tendermint.CreateTestHeader(chainID, 6, suite.valSet, bothValSet, signers),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         "gaiamainnet",
			},
			func(ev *tendermint.Evidence) error { return nil },
			false,
		},
		{
			"same block id",
			tendermint.Evidence{
				Header1:          suite.header,
				Header2:          suite.header,
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         "gaiamainnet",
			},
			func(ev *tendermint.Evidence) error { return nil },
			false,
		},
		{
			"header 1 doesn't have 2/3 majority",
			tendermint.Evidence{
				Header1:          tendermint.CreateTestHeader(chainID, height, bothValSet, bothValSet, bothSigners),
				Header2:          suite.header,
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         "gaiamainnet",
			},
			func(ev *tendermint.Evidence) error {
				// voteSet contains only altVal which is less than 2/3 of total power (height/1height)
				wrongVoteSet := tmtypes.NewVoteSet(chainID, ev.Header1.Height, 1, tmtypes.PrecommitType, altValSet)
				var err error
				ev.Header1.Commit, err = tmtypes.MakeCommit(ev.Header1.Commit.BlockID, ev.Header2.Height, ev.Header1.Commit.Round, wrongVoteSet, altSigners)
				return err
			},
			false,
		},
		{
			"header 2 doesn't have 2/3 majority",
			tendermint.Evidence{
				Header1:          suite.header,
				Header2:          tendermint.CreateTestHeader(chainID, height, bothValSet, bothValSet, bothSigners),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         "gaiamainnet",
			},
			func(ev *tendermint.Evidence) error {
				// voteSet contains only altVal which is less than 2/3 of total power (height/1height)
				wrongVoteSet := tmtypes.NewVoteSet(chainID, ev.Header2.Height, 1, tmtypes.PrecommitType, altValSet)
				var err error
				ev.Header2.Commit, err = tmtypes.MakeCommit(ev.Header2.Commit.BlockID, ev.Header2.Height, ev.Header2.Commit.Round, wrongVoteSet, altSigners)
				return err
			},
			false,
		},
		{
			"validators sign off on wrong commit",
			tendermint.Evidence{
				Header1:          suite.header,
				Header2:          tendermint.CreateTestHeader(chainID, height, bothValSet, bothValSet, bothSigners),
				FromValidatorSet: bothValSet,
				ChainID:          chainID,
				ClientID:         "gaiamainnet",
			},
			func(ev *tendermint.Evidence) error {
				ev.Header2.Commit.BlockID = tendermint.MakeBlockID(tmhash.Sum([]byte("other_hash")), 3, tmhash.Sum([]byte("other_partset")))
				return nil
			},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.malleateEvidence(&tc.evidence)
		suite.Require().NoError(err)

		if tc.expPass {
			suite.Require().NoError(tc.evidence.ValidateBasic(), "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(tc.evidence.ValidateBasic(), "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
