package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simutil "github.com/cosmos/cosmos-sdk/x/simulation/util"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

// Simulation parameter constants
const (
	SignedBlocksWindow      = "signed_blocks_window"
	MinSignedPerWindow      = "min_signed_per_window"
	DowntimeJailDuration    = "downtime_jail_duration"
	SlashFractionDoubleSign = "slash_fraction_double_sign"
	SlashFractionDowntime   = "slash_fraction_downtime"
)

// GenSlashingGenesisState generates a random GenesisState for slashing
func GenSlashingGenesisState(
	cdc *codec.Codec, r *rand.Rand, maxEvidenceAge time.Duration,
	ap simulation.AppParams, genesisState map[string]json.RawMessage,
) {

	var signedBlocksWindow int64
	ap.GetOrGenerate(cdc, SignedBlocksWindow, &signedBlocksWindow, r,
		func(r *rand.Rand) {
			signedBlocksWindow = int64(simutil.RandIntBetween(r, 10, 1000))
		})

	var minSignedPerWindow sdk.Dec
	ap.GetOrGenerate(cdc, MinSignedPerWindow, &minSignedPerWindow, r,
		func(r *rand.Rand) {
			minSignedPerWindow = sdk.NewDecWithPrec(int64(r.Intn(10)), 1)
		})

	var downtimeJailDuration time.Duration
	ap.GetOrGenerate(cdc, DowntimeJailDuration, &downtimeJailDuration, r,
		func(r *rand.Rand) {
			downtimeJailDuration = time.Duration(simutil.RandIntBetween(r, 60, 60*60*24)) * time.Second
		})

	var slashFractionDoubleSign sdk.Dec
	ap.GetOrGenerate(cdc, SlashFractionDoubleSign, &slashFractionDoubleSign, r,
		func(r *rand.Rand) {
			slashFractionDoubleSign = sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(50) + 1)))
		})

	var slashFractionDowntime sdk.Dec
	ap.GetOrGenerate(cdc, SlashFractionDowntime, &slashFractionDowntime, r,
		func(r *rand.Rand) {
			slashFractionDowntime = sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(200) + 1)))
		})

	params := slashing.NewParams(maxEvidenceAge, signedBlocksWindow, minSignedPerWindow,
		downtimeJailDuration, slashFractionDoubleSign, slashFractionDowntime)
	slashingGenesis := slashing.NewGenesisState(params, nil, nil)

	fmt.Printf("Selected randomly generated slashing parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, slashingGenesis.Params))
	genesisState[slashing.ModuleName] = cdc.MustMarshalJSON(slashingGenesis)
}
