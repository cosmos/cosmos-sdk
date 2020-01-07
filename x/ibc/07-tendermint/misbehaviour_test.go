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

	committer := Committer{
		ValidatorSet:   suite.valSet,
		Height:         3,
		NextValSetHash: suite.valSet.Hash(),
	}

	testCases := []struct {
		name     string
		evidence Evidence
		expErr   bool
	}{
		{
			"trusting period misbehavior should pass",
			Evidence{
				Header1:  MakeHeader("gaia", 5, bothValSet, suite.valSet, bothSigners),
				Header2:  MakeHeader("gaia", 5, bothValSet, bothValSet, bothSigners),
				ChainID:  "gaia",
				ClientID: "gaiamainnet",
			},
			false,
		},
		{
			"first valset has too much change",
			Evidence{
				Header1:  MakeHeader("gaia", 5, altValSet, bothValSet, altSigners),
				Header2:  MakeHeader("gaia", 5, bothValSet, bothValSet, bothSigners),
				ChainID:  "gaia",
				ClientID: "gaiamainnet",
			},
			true,
		},
		{
			"second valset has too much change",
			Evidence{
				Header1:  MakeHeader("gaia", 5, bothValSet, bothValSet, bothSigners),
				Header2:  MakeHeader("gaia", 5, altValSet, bothValSet, altSigners),
				ChainID:  "gaia",
				ClientID: "gaiamainnet",
			},
			true,
		},
		{
			"both valsets have too much change",
			Evidence{
				Header1:  MakeHeader("gaia", 5, altValSet, altValSet, altSigners),
				Header2:  MakeHeader("gaia", 5, altValSet, bothValSet, altSigners),
				ChainID:  "gaia",
				ClientID: "gaiamainnet",
			},
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc // pin for scopelint
		suite.Run(tc.name, func() {
			err := CheckMisbehaviour(committer, tc.evidence)
			if tc.expErr {
				suite.Error(err, "CheckMisbehaviour passed unexpectedly")
			} else {
				suite.NoError(err, "CheckMisbehaviour failed unexpectedly: %v", err)
			}
		})
	}
}
