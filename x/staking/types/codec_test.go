package types_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/staking/client/rest"
	"github.com/stretchr/testify/require"
)

func TestCodec(t *testing.T) {
	jsonData := `
{
      "base_req": {
        "from": "cosmos1j4heycy25ersvld236f3ckpn9avjjt4p3tmg23",
        "memo": "Delegation",
        "chain_id": "cosmoshub-3",
        "account_number": "30429",
        "sequence": "11",
        "gas": "500000",
        "gas_adjustment": "1.15",
        "fees": [
          {
            "denom": "uatom",
            "amount": "100000"
          }
        ],
        "simulate": true
      },
      "delegator_address": "cosmos1j4heycy25ersvld236f3ckpn9avjjt4p3tmg23",
      "validator_address": "cosmosvaloper1we6knm8qartmmh2r0qfpsz6pq0s7emv3e0meuw",
      "delegation": {
        "amount": "10000",
        "denom": "uatom"
      }

}
`
	var req rest.DelegateRequest

	err := types.ModuleCdc.UnmarshalJSON([]byte(jsonData), &req)
	require.NoError(t, err)

	fmt.Printf("%v", req.Amount)
}
