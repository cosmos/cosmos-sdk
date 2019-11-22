package tendermint

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"
)

func (suite *TendermintTestSuite) TestValidateBasic() {
	altPrivVal := tmtypes.NewMockPV()
	altVal := tmtypes.NewValidator(altPrivVal.GetPubKey(), 4)

	altValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	wrongValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	signers := []tmtypes.PrivValidator{suite.privVal}
	var bothSigners []tmtypes.PrivValidator
	if bytes.Compare(altPrivVal.GetPubKey().Address(), suite.privVal.GetPubKey().Address()) == -1 {
		bothSigners = []tmtypes.PrivValidator{altPrivVal, suite.privVal}
	} else {
		bothSigners = []tmtypes.PrivValidator{suite.privVal, altPrivVal}
	}

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	testCases := []struct {
		name             string
		evidence         Evidence
		malleateEvidence func(ev *Evidence)
		expErr           bool
	}{
		{
			"valid evidence",
			Evidence{
				Header1: suite.header,
				Header2: MakeHeader("gaia", 4, suite.valSet, altValSet, signers),
				ChainID: "gaia",
			},
			func(ev *Evidence) {},
			false,
		},
		{
			"wrong chainID on header1",
			Evidence{
				Header1: suite.header,
				Header2: MakeHeader("ethermint", 4, suite.valSet, altValSet, signers),
				ChainID: "ethermint",
			},
			func(ev *Evidence) {},
			true,
		},
		{
			"wrong chainID on header2",
			Evidence{
				Header1: suite.header,
				Header2: MakeHeader("ethermint", 4, suite.valSet, altValSet, signers),
				ChainID: "gaia",
			},
			func(ev *Evidence) {},
			true,
		},
		{
			"mismatched heights",
			Evidence{
				Header1: suite.header,
				Header2: MakeHeader("gaia", 6, suite.valSet, altValSet, signers),
				ChainID: "gaia",
			},
			func(ev *Evidence) {},
			true,
		},
		{
			"mismatched heights",
			Evidence{
				Header1: suite.header,
				Header2: MakeHeader("gaia", 6, suite.valSet, altValSet, signers),
				ChainID: "gaia",
			},
			func(ev *Evidence) {},
			true,
		},
		{
			"same block id",
			Evidence{
				Header1: suite.header,
				Header2: suite.header,
				ChainID: "gaia",
			},
			func(ev *Evidence) {},
			true,
		},
		{
			"header doesn't have 2/3 majority",
			Evidence{
				Header1: suite.header,
				Header2: MakeHeader("gaia", 4, altValSet, altValSet, bothSigners),
				ChainID: "gaia",
			},
			func(ev *Evidence) {
				fmt.Printf("ValidatorSet: %v\n\n", wrongValSet)
				fmt.Printf("Signer: %v\n\n", altSigners)
				wrongVoteSet := tmtypes.NewVoteSet("gaia", ev.Header2.Height, 1, tmtypes.PrecommitType, wrongValSet)
				var err error
				ev.Header2.Commit, err = tmtypes.MakeCommit(ev.Header2.Commit.BlockID, ev.Header2.Height, ev.Header2.Commit.Round(), wrongVoteSet, altSigners)
				if err != nil {
					panic(err)
				}
			},
			true,
		},
		{
			"validators sign off on wrong commit",
			Evidence{
				Header1: suite.header,
				Header2: MakeHeader("gaia", 4, altValSet, altValSet, bothSigners),
				ChainID: "gaia",
			},
			func(ev *Evidence) {
				ev.Header2.Commit.BlockID = makeBlockID(tmhash.Sum([]byte("other_hash")), 3, tmhash.Sum([]byte("other_partset")))
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// reset suite for each subtest
			suite.SetupTest()
			fmt.Println(tc.name)

			tc.malleateEvidence(&tc.evidence)

			err := tc.evidence.ValidateBasic()

			if tc.expErr {
				require.NotNil(suite.T(), err, "ValidateBasic did not throw error for invalid evidence")
			} else {
				require.Nil(suite.T(), err, "ValidateBasic returned error on valid evidence: %s", err)
			}
		})
	}
}
