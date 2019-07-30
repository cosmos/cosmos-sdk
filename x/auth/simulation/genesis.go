package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// GenAuthGenesisState generates a random GenesisState for auth
func GenAuthGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	authGenesis := auth.NewGenesisState(
		auth.NewParams(
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.MaxMemoChars, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.MaxMemoChars](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.TxSigLimit, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TxSigLimit](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.TxSizeCostPerByte, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.TxSizeCostPerByte](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.SigVerifyCostED25519, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SigVerifyCostED25519](r).(uint64)
					})
				return v
			}(r),
			func(r *rand.Rand) uint64 {
				var v uint64
				ap.GetOrGenerate(cdc, simulation.SigVerifyCostSECP256K1, &v, r,
					func(r *rand.Rand) {
						v = simulation.ModuleParamSimulator[simulation.SigVerifyCostSECP256K1](r).(uint64)
					})
				return v
			}(r),
		),
	)

	fmt.Printf("Selected randomly generated auth parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, authGenesis.Params))
	genesisState[auth.ModuleName] = cdc.MustMarshalJSON(authGenesis)
}
