package tendermint

import (
	"bytes"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

func (suite *TendermintTestSuite) TestMisbehaviourValidateBasic() {
	altPrivVal := tmtypes.NewMockPV()
	altVal := tmtypes.NewValidator(altPrivVal.GetPubKey(), 4)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))

	signers := []tmtypes.PrivValidator{suite.privVal}
	testCases := []struct {
		name     string
		evidence *Evidence
		clientID string
		expErr   bool
	}{
		{
			"valid misbehavior",
			&Evidence{
				Header1: suite.header,
				Header2: MakeHeader("gaia", 4, suite.valSet, bothValSet, signers),
				ChainID: "gaia",
			},
			"gaiamainnet",
			false,
		},
		{
			"nil evidence",
			nil,
			"gaiamainnet",
			true,
		},
		{
			"invalid evidence",
			&Evidence{
				Header1: suite.header,
				Header2: suite.header,
				ChainID: "gaia",
			},
			"gaiamainnet",
			true,
		},
		{
			"invalid ClientID",
			&Evidence{
				Header1: MakeHeader("gaia123??", 4, suite.valSet, suite.valSet, signers),
				Header2: MakeHeader("gaia123?", 4, suite.valSet, suite.valSet, signers),
				ChainID: "gaia123?",
			},
			"gaia123?",
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			misbehaviour := Misbehaviour{
				Evidence: tc.evidence,
				ClientID: tc.clientID,
			}

			err := misbehaviour.ValidateBasic()

			if tc.expErr {
				require.NotNil(suite.T(), err, "Invalid Misbehaviour passed ValidateBasic")
			} else {
				require.Nil(suite.T(), err, "Valid Misbehaviour failed ValidateBasic: %v", err)
			}
		})
	}
}

func (suite *TendermintTestSuite) TestCheckMisbehaviour() {
	altPrivVal := tmtypes.NewMockPV()
	altVal := tmtypes.NewValidator(altPrivVal.GetPubKey(), 4)

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
		name      string
		evidence  *Evidence
		clientID  string
		committer Committer
		expErr    bool
	}{
		{
			"misbehavior should pass",
			&Evidence{
				Header1: MakeHeader("gaia", 4, bothValSet, suite.valSet, bothSigners),
				Header2: MakeHeader("gaia", 4, bothValSet, bothValSet, signers),
				ChainID: "gaia",
			},
			"gaiamainnet",
			Committer{suite.valSet},
			false,
		},
		{
			"first valset has too much change",
			&Evidence{
				Header1: MakeHeader("gaia", 4, altValSet, bothValSet, altSigners),
				Header2: MakeHeader("gaia", 4, bothValSet, bothValSet, bothSigners),
				ChainID: "gaia",
			},
			"gaiamainnet",
			Committer{suite.valSet},
			true,
		},
		{
			"second valset has too much change",
			&Evidence{
				Header1: MakeHeader("gaia", 4, bothValSet, bothValSet, bothSigners),
				Header2: MakeHeader("gaia", 4, altValSet, bothValSet, altSigners),
				ChainID: "gaia",
			},
			"gaiamainnet",
			Committer{suite.valSet},
			true,
		},
		{
			"both valsets have too much change",
			&Evidence{
				Header1: MakeHeader("gaia", 4, altValSet, altValSet, altSigners),
				Header2: MakeHeader("gaia", 4, altValSet, bothValSet, altSigners),
				ChainID: "gaia",
			},
			"gaiamainnet",
			Committer{suite.valSet},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			misbehaviour := Misbehaviour{
				Evidence: tc.evidence,
				ClientID: tc.clientID,
			}

			err := CheckMisbehaviour(tc.committer, misbehaviour)

			if tc.expErr {
				require.NotNil(suite.T(), err, "CheckMisbehaviour passed unexpectedly")
			} else {
				require.Nil(suite.T(), err, "CheckMisbehaviour failed unexpectedly: %v", err)
			}
		})
	}
}
