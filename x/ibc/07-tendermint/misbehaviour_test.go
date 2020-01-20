package tendermint

import (
	"bytes"

	tmtypes "github.com/tendermint/tendermint/types"
)

func (suite *TendermintTestSuite) TestCheckMisbehaviour() {
	altPrivVal := tmtypes.NewMockPV()
	altVal := tmtypes.NewValidator(altPrivVal.GetPubKey(), 4)

	// Create bothValSet with both suite validator and altVal
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	// Create signer array and ensure it is in same order as bothValSet
	var bothSigners []tmtypes.PrivValidator
	if bytes.Compare(altPrivVal.GetPubKey().Address(), suite.privVal.GetPubKey().Address()) == -1 {
		bothSigners = []tmtypes.PrivValidator{altPrivVal, suite.privVal}
	} else {
		bothSigners = []tmtypes.PrivValidator{suite.privVal, altPrivVal}
	}

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	testCases := []struct {
		name           string
		clientState    ClientState
		consensusState ConsensusState
		evidence       Evidence
		height         uint64
		expErr         bool
	}{
		{
			"trusting period misbehavior should pass",
			ClientState{},
			ConsensusState{},
			Evidence{
				Header1:  CreateTestHeader("gaia", 5, bothValSet, suite.valSet, bothSigners),
				Header2:  CreateTestHeader("gaia", 5, bothValSet, bothValSet, bothSigners),
				ChainID:  "gaia",
				ClientID: "gaiamainnet",
			},
			5,
			false,
		},
		{
			"first valset has too much change",
			ClientState{},
			ConsensusState{},
			Evidence{
				Header1:  CreateTestHeader("gaia", 5, altValSet, bothValSet, altSigners),
				Header2:  CreateTestHeader("gaia", 5, bothValSet, bothValSet, bothSigners),
				ChainID:  "gaia",
				ClientID: "gaiamainnet",
			},
			5,
			true,
		},
		{
			"second valset has too much change",
			ClientState{},
			ConsensusState{},
			Evidence{
				Header1:  CreateTestHeader("gaia", 5, bothValSet, bothValSet, bothSigners),
				Header2:  CreateTestHeader("gaia", 5, altValSet, bothValSet, altSigners),
				ChainID:  "gaia",
				ClientID: "gaiamainnet",
			},
			5,
			true,
		},
		{
			"both valsets have too much change",
			ClientState{},
			ConsensusState{},
			Evidence{
				Header1:  CreateTestHeader("gaia", 5, altValSet, altValSet, altSigners),
				Header2:  CreateTestHeader("gaia", 5, altValSet, bothValSet, altSigners),
				ChainID:  "gaia",
				ClientID: "gaiamainnet",
			},
			5,
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc // pin for scopelint
		suite.Run(tc.name, func() {
			err := checkMisbehaviour(tc.clientState, tc.consensusState, tc.evidence, tc.height)
			if tc.expErr {
				suite.Error(err, "CheckMisbehaviour passed unexpectedly")
			} else {
				suite.NoError(err, "CheckMisbehaviour failed unexpectedly: %v", err)
			}
		})
	}
}
