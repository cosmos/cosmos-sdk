package safety_test

import (
	"encoding/json"
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/simapp"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

// TestGenesisExportImportRoundTrip trips several patterns, exports the whole
// app, starts a fresh app from the export, and checks that x/safety state,
// enforcement, and a second export all match.
func TestGenesisExportImportRoundTrip(t *testing.T) {
	f := newFixture(t, nil, nil)
	f.setParams(f.operator.String(), "/cosmos.bank.*", "/cosmos.staking.*")
	f.mustDeliver(f.operator, safetytypes.NewMsgTrip(f.operator.String(), msgSendURL, "/cosmos.staking.*"))
	wantParams := f.params()
	wantDisabled := f.disabled()
	require.ElementsMatch(t, []string{msgSendURL, "/cosmos.staking.*"}, wantDisabled)

	exported, err := f.app.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)

	var appState map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(exported.AppState, &appState))
	var gs safetytypes.GenesisState
	require.NoError(t, f.app.AppCodec().UnmarshalJSON(appState[safetytypes.ModuleName], &gs))
	require.Equal(t, wantParams, gs.Params)
	require.ElementsMatch(t, wantDisabled, gs.DisabledPatterns)
	require.NoError(t, gs.Validate(f.app.AccountKeeper.AddressCodec()))

	// the whole exported app state passes validate-genesis
	require.NoError(t, f.app.BasicModuleManager.ValidateGenesis(f.app.AppCodec(), f.app.TxConfig(), appState))

	validators := make([]abci.ValidatorUpdate, len(exported.Validators))
	for i, v := range exported.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(v.PubKey)
		require.NoError(t, err)
		protoPk, err := cryptocodec.ToCmtProtoPublicKey(pk)
		require.NoError(t, err)
		validators[i] = abci.ValidatorUpdate{PubKey: protoPk, Power: v.Power}
	}

	app2 := simapp.NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
	_, err = app2.InitChain(&abci.RequestInitChain{
		Time:            f.now,
		Validators:      validators,
		ConsensusParams: &exported.ConsensusParams,
		AppStateBytes:   exported.AppState,
		InitialHeight:   exported.Height,
	})
	require.NoError(t, err)

	g := &fixture{t: t, app: app2, valSet: f.valSet, now: f.now, initialHeight: exported.Height, voter: f.voter, user: f.user, operator: f.operator, other: f.other}
	g.finalize()

	require.Equal(t, wantParams, g.params())
	require.ElementsMatch(t, wantDisabled, g.disabled())
	require.False(t, g.isAllowed(msgSendURL))
	require.False(t, g.isAllowed(msgDelegateURL))
	require.True(t, g.isAllowed("/cosmos.bank.v1beta1.MsgMultiSend"))

	// the imported trip still blocks
	g.requireRejected(g.user, safetytypes.ErrMsgDisabled, send(g.user, g.other))
	valAddr := sdk.ValAddress(g.valSet.Validators[0].Address).String()
	g.requireRejected(g.user, safetytypes.ErrMsgDisabled, stakingDelegate(g.user, valAddr))

	// a second export of the imported app matches the first
	reexported, err := app2.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	var appState2 map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(reexported.AppState, &appState2))
	require.JSONEq(t, string(appState[safetytypes.ModuleName]), string(appState2[safetytypes.ModuleName]))

	// and the imported operator resets as before
	g.mustDeliver(g.operator, safetytypes.NewMsgReset(g.operator.String(), msgSendURL, "/cosmos.staking.*"))
	require.Empty(t, g.disabled())
	g.mustDeliver(g.user, send(g.user, g.other))
}

// TestGenesisDisabledProtectedIsInert checks that a genesis disabled list
// naming protected msgs, which the operator could never trip, imports but
// cannot block them, so reset lockout is impossible even via genesis.
func TestGenesisDisabledProtectedIsInert(t *testing.T) {
	disabled := []string{"/cosmos.gov.v1.MsgSubmitProposal", "/cosmos.gov.*", "/cosmos.safety.v1.MsgReset", "/cosmos.safety.*", msgSendURL}
	f := newFixture(t, nil, func(gs simapp.GenesisState) {
		var sgs safetytypes.GenesisState
		require.NoError(t, json.Unmarshal(gs[safetytypes.ModuleName], &sgs))
		sgs.DisabledPatterns = disabled
		bz, err := json.Marshal(&sgs)
		require.NoError(t, err)
		gs[safetytypes.ModuleName] = bz
	})
	require.ElementsMatch(t, disabled, f.disabled())
	require.True(t, f.isAllowed("/cosmos.gov.v1.MsgSubmitProposal"))
	require.True(t, f.isAllowed("/cosmos.safety.v1.MsgReset"))
	require.False(t, f.isAllowed(msgSendURL))

	f.mustPass(safetytypes.NewMsgReset(govAddr, msgSendURL))
	f.mustDeliver(f.user, send(f.user, f.other))
}

// TestGenesisInvalidRejected checks that InitChain fails on an invalid
// x/safety genesis.
func TestGenesisInvalidRejected(t *testing.T) {
	for name, mutate := range map[string]func(*safetytypes.GenesisState){
		"malformed pattern": func(gs *safetytypes.GenesisState) { gs.DisabledPatterns = []string{"cosmos.bank.v1beta1.MsgSend"} },
		"duplicate pattern": func(gs *safetytypes.GenesisState) { gs.DisabledPatterns = []string{msgSendURL, msgSendURL} },
		"empty authority":   func(gs *safetytypes.GenesisState) { gs.Params.Authority = "" },
		"bad operator":      func(gs *safetytypes.GenesisState) { gs.Params.Operator = "nope" },
	} {
		t.Run(name, func(t *testing.T) {
			app := simapp.NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.NewAppOptionsWithFlagHome(t.TempDir()))
			genesis := app.DefaultGenesis()
			var sgs safetytypes.GenesisState
			require.NoError(t, app.AppCodec().UnmarshalJSON(genesis[safetytypes.ModuleName], &sgs))
			mutate(&sgs)
			genesis[safetytypes.ModuleName] = app.AppCodec().MustMarshalJSON(&sgs)

			// validate-genesis rejects it
			require.Error(t, app.BasicModuleManager.ValidateGenesis(app.AppCodec(), app.TxConfig(), genesis))

			// and so does InitChain, by error or panic
			stateBytes, err := json.Marshal(genesis)
			require.NoError(t, err)
			err = func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("panic: %v", r)
					}
				}()
				_, err = app.InitChain(&abci.RequestInitChain{
					Validators:      []abci.ValidatorUpdate{},
					ConsensusParams: defaultConsensusParams(),
					AppStateBytes:   stateBytes,
				})
				return err
			}()
			require.Error(t, err)
		})
	}
}
