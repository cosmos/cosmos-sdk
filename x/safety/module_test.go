package safety_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/simsx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/safety"
	"github.com/cosmos/cosmos-sdk/x/safety/keeper"
	"github.com/cosmos/cosmos-sdk/x/safety/simulation"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

func TestModuleGenesis(t *testing.T) {
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test")).Ctx
	encCfg := moduletestutil.MakeTestEncodingConfig()
	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	gov := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	k := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), addresscodec.NewBech32Codec("cosmos"), encCfg.InterfaceRegistry, gov)
	am := safety.NewAppModule(k)

	def := am.DefaultGenesis(encCfg.Codec)
	require.NoError(t, am.ValidateGenesis(encCfg.Codec, nil, def))
	var gs types.GenesisState
	require.NoError(t, encCfg.Codec.UnmarshalJSON(def, &gs))
	require.Equal(t, gov, gs.Params.Authority)
	require.Empty(t, gs.Params.Operator)
	require.Empty(t, gs.Params.Scope, "default scope must be empty")
	require.Empty(t, gs.DisabledPatterns)

	gs.DisabledPatterns = []string{"/cosmos.bank.*", "/cosmos.staking.v1beta1.MsgDelegate"}
	bz := encCfg.Codec.MustMarshalJSON(&gs)
	require.NoError(t, am.ValidateGenesis(encCfg.Codec, nil, bz))
	am.InitGenesis(ctx, encCfg.Codec, bz)
	exported := am.ExportGenesis(ctx, encCfg.Codec)
	require.JSONEq(t, string(bz), string(exported))

	bad := json.RawMessage(`{"params":{"authority":"` + gov + `","scope":["*"]},"disabled_patterns":["oops"]}`)
	require.ErrorIs(t, am.ValidateGenesis(encCfg.Codec, nil, bad), types.ErrInvalidPattern)
	require.Panics(t, func() { am.InitGenesis(ctx, encCfg.Codec, bad) })
}

// A zero-value AppModule, as a static BasicManager stores it, emits a default
// genesis that validates and validates genesis without a keeper.
func TestZeroModuleGenesis(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	am := safety.AppModule{}
	gov := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	def := am.DefaultGenesis(encCfg.Codec)
	var gs types.GenesisState
	require.NoError(t, encCfg.Codec.UnmarshalJSON(def, &gs))
	require.Equal(t, gov, gs.Params.Authority)
	require.NoError(t, am.ValidateGenesis(encCfg.Codec, nil, def))

	gs.Params.Operator = sdk.AccAddress("zero_operator_______").String()
	gs.Params.Scope = []string{"/cosmos.bank.*"}
	gs.DisabledPatterns = []string{"/cosmos.bank.*"}
	require.NoError(t, am.ValidateGenesis(encCfg.Codec, nil, encCfg.Codec.MustMarshalJSON(&gs)))

	gs.Params.Operator = "cosmos1bad"
	require.ErrorIs(t, am.ValidateGenesis(encCfg.Codec, nil, encCfg.Codec.MustMarshalJSON(&gs)), sdkerrors.ErrInvalidAddress)
}

type weights []uint32

func (w *weights) Add(weight uint32, _ simsx.SimMsgFactoryX) { *w = append(*w, weight) }

// The simulation hooks generate a valid genesis under the keeper's default
// authority, decode the module store, register one gov proposal msg and
// no random operations.
func TestModuleSimulation(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	authority := sdk.AccAddress("sim_authority_______").String()
	k := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(storetypes.NewKVStoreKey(types.StoreKey)), addresscodec.NewBech32Codec("cosmos"), encCfg.InterfaceRegistry, authority)
	am := safety.NewAppModule(k)

	r := rand.New(rand.NewSource(1))
	simState := module.SimulationState{
		Cdc:      encCfg.Codec,
		Rand:     r,
		GenState: make(map[string]json.RawMessage),
		Accounts: simtypes.RandomAccounts(r, 3),
	}
	am.GenerateGenesisState(&simState)
	require.NoError(t, am.ValidateGenesis(encCfg.Codec, nil, simState.GenState[types.ModuleName]))
	var gs types.GenesisState
	encCfg.Codec.MustUnmarshalJSON(simState.GenState[types.ModuleName], &gs)
	require.Equal(t, authority, gs.Params.Authority)

	sdr := make(simtypes.StoreDecoderRegistry)
	am.RegisterStoreDecoder(sdr)
	require.Contains(t, sdr, types.StoreKey)

	var reg weights
	am.ProposalMsgsX(simsx.WeightSourceFn(func(_ string, def uint32) uint32 { return def }), &reg)
	require.Equal(t, weights{uint32(simulation.DefaultWeightMsgUpdateParams)}, reg)

	require.Empty(t, am.WeightedOperations(simState))
}
