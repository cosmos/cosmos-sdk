package types_test

import (
	"testing"

	types2 "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/staking/client/rest"
	"github.com/stretchr/testify/require"
)

func TestCodec_BadRequest(t *testing.T) {
	badRequestAmount := `
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

	err := types.ModuleCdc.UnmarshalJSON([]byte(badRequestAmount), &req)
	require.NoError(t, err)
	require.Equal(t, types2.Coin{Denom: "", Amount: types2.Int{}}, req.Amount)
}

func TestCodec_GoodRequest(t *testing.T) {
	badRequestAmount := `
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
      "amount": {
        "amount": "10000",
        "denom": "uatom"
      }

}
`
	var req rest.DelegateRequest

	err := types.ModuleCdc.UnmarshalJSON([]byte(badRequestAmount), &req)
	require.NoError(t, err)
	require.Equal(t, types2.Coin{Denom: "uatom", Amount: types2.NewInt(10000)}, req.Amount)

	// Good request
}
