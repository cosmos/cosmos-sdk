package v036

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/x/genutil"
)

var basic034Gov = []byte(`
    {
      "starting_proposal_id": "2",
      "deposits": [
        {
          "proposal_id": "1",
          "deposit": {
            "depositor": "cosmos1grgelyng2v6v3t8z87wu3sxgt9m5s03xvslewd",
            "proposal_id": "1",
            "amount": [
              {
                "denom": "uatom",
                "amount": "512000000"
              }
            ]
          }
        }
      ],
      "votes" : [
        {
          "proposal_id": "1",
          "vote": {
            "voter": "cosmos1lktjhnzkpkz3ehrg8psvmwhafg56kfss5597tg",
            "proposal_id": "1",
            "option": "Yes"
          }
        }
      ],
      "proposals": [
        {
          "proposal_content": {
            "type": "gov/TextProposal",
            "value": {
              "title": "test",
              "description": "test"
            }
          },
          "proposal_id": "1",
          "proposal_status": "Passed",
          "final_tally_result": {
            "yes": "1",
            "abstain": "0",
            "no": "0",
            "no_with_veto": "0"
          },
          "submit_time": "2019-05-03T21:08:25.443199036Z",
          "deposit_end_time": "2019-05-17T21:08:25.443199036Z",
          "total_deposit": [
            {
              "denom": "uatom",
              "amount": "512000000"
            }
          ],
          "voting_start_time": "2019-05-04T16:02:33.24680295Z",
          "voting_end_time": "2019-05-18T16:02:33.24680295Z"
        }
      ],
      "deposit_params": {
        "min_deposit": [
          {
            "denom": "uatom",
            "amount": "512000000"
          }
        ],
        "max_deposit_period": "1209600000000000"
      },
      "voting_params": {
        "voting_period": "1209600000000000"
      },
      "tally_params": {
        "quorum": "0.400000000000000000",
        "threshold": "0.500000000000000000",
        "veto": "0.334000000000000000"
      }
    }
`)

func TestDummyGenesis(t *testing.T) {
	genesisDummy := genutil.AppMap{
		"foo": {},
		"bar": []byte(`{"custom": "module"}`),
	}
	cdc := amino.NewCodec()
	migratedDummy := Migrate(genesisDummy, cdc)

	// We should not touch custom modules in the map
	require.Equal(t, genesisDummy["foo"], migratedDummy["foo"])
	require.Equal(t, genesisDummy["bar"], migratedDummy["bar"])
}

func TestGovGenesis(t *testing.T) {
	genesis := genutil.AppMap{
		"gov": basic034Gov,
	}
	cdc := amino.NewCodec()

	require.NotPanics(t, func() { Migrate(genesis, cdc) })
}
