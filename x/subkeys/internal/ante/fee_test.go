package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/subkeys/internal/ante"
	"github.com/cosmos/cosmos-sdk/x/subkeys/internal/types"
	// delTypes "github.com/cosmos/cosmos-sdk/x/subkeys/internal/types"
)

func TestDeductFeesNoDelegation(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)

	// this just tests our handler
	dfd := ante.NewDeductDelegatedFeeDecorator(app.AccountKeeper, app.SupplyKeeper, app.DelegationKeeper)
	ourAnteHandler := sdk.ChainAnteDecorators(dfd)

	// this tests the whole stack
	anteHandlerStack := ante.NewAnteHandler(app.AccountKeeper, app.SupplyKeeper, app.DelegationKeeper, SigGasNoConsumer)

	// keys and addresses
	priv1, _, addr1 := authtypes.KeyTestPubAddr()
	priv2, _, addr2 := authtypes.KeyTestPubAddr()
	priv3, _, addr3 := authtypes.KeyTestPubAddr()
	priv4, _, addr4 := authtypes.KeyTestPubAddr()

	// Set addr1 with insufficient funds
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins([]sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(10))})
	app.AccountKeeper.SetAccount(ctx, acc1)

	// Set addr2 with more funds
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins([]sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(99999))})
	app.AccountKeeper.SetAccount(ctx, acc2)

	// Set delegation from addr2 to addr3 (plenty to pay)
	err := app.DelegationKeeper.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr2,
		Grantee: addr3,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 500)),
		},
	})
	require.NoError(t, err)

	// Set low delegation from addr2 to addr4 (delegation will reject)
	err = app.DelegationKeeper.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr2,
		Grantee: addr4,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 20)),
		},
	})
	require.NoError(t, err)

	// Set delegation from addr1 to addr4 (cannot cover this )
	err = app.DelegationKeeper.DelegateFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr2,
		Grantee: addr3,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 500)),
		},
	})
	require.NoError(t, err)

	cases := map[string]struct {
		signerKey  crypto.PrivKey
		signer     sdk.AccAddress
		feeAccount sdk.AccAddress
		handler    sdk.AnteHandler
		fee        int64
		valid      bool
	}{
		"paying with low funds (only ours)": {
			signerKey: priv1,
			signer:    addr1,
			fee:       50,
			handler:   ourAnteHandler,
			valid:     false,
		},
		"paying with good funds (only ours)": {
			signerKey: priv2,
			signer:    addr2,
			fee:       50,
			handler:   ourAnteHandler,
			valid:     true,
		},
		"paying with no account (only ours)": {
			signerKey: priv3,
			signer:    addr3,
			fee:       1,
			handler:   ourAnteHandler,
			valid:     false,
		},
		"no fee with real account (only ours)": {
			signerKey: priv1,
			signer:    addr1,
			fee:       0,
			handler:   ourAnteHandler,
			valid:     true,
		},
		"no fee with no account (only ours)": {
			signerKey: priv4,
			signer:    addr4,
			fee:       0,
			handler:   ourAnteHandler,
			valid:     false,
		},
		"valid delegation without account (only ours)": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr2,
			fee:        50,
			handler:    ourAnteHandler,
			valid:      true,
		},
		"no delegation (only ours)": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr1,
			fee:        2,
			handler:    ourAnteHandler,
			valid:      false,
		},
		"allowance smaller than requested fee (only ours)": {
			signerKey:  priv4,
			signer:     addr4,
			feeAccount: addr2,
			fee:        50,
			handler:    ourAnteHandler,
			valid:      false,
		},
		"granter cannot cover allowed delegation (only ours)": {
			signerKey:  priv4,
			signer:     addr4,
			feeAccount: addr1,
			fee:        50,
			handler:    ourAnteHandler,
			valid:      false,
		},

		"paying with low funds (whole stack)": {
			signerKey: priv1,
			signer:    addr1,
			fee:       50,
			handler:   anteHandlerStack,
			valid:     false,
		},
		"paying with good funds (whole stack)": {
			signerKey: priv2,
			signer:    addr2,
			fee:       50,
			handler:   anteHandlerStack,
			valid:     true,
		},
		"paying with no account (whole stack)": {
			signerKey: priv3,
			signer:    addr3,
			fee:       1,
			handler:   anteHandlerStack,
			valid:     false,
		},
		"no fee with real account (whole stack)": {
			signerKey: priv1,
			signer:    addr1,
			fee:       0,
			handler:   anteHandlerStack,
			valid:     true,
		},
		"no fee with no account (whole stack)": {
			signerKey: priv4,
			signer:    addr4,
			fee:       0,
			handler:   anteHandlerStack,
			valid:     false,
		},
		"valid delegation without account (whole stack)": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr2,
			fee:        50,
			handler:    anteHandlerStack,
			valid:      true,
		},
		"no delegation (whole stack)": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr1,
			fee:        2,
			handler:    anteHandlerStack,
			valid:      false,
		},
		"allowance smaller than requested fee (whole stack)": {
			signerKey:  priv4,
			signer:     addr4,
			feeAccount: addr2,
			fee:        50,
			handler:    anteHandlerStack,
			valid:      false,
		},
		"granter cannot cover allowed delegation (whole stack)": {
			signerKey:  priv4,
			signer:     addr4,
			feeAccount: addr1,
			fee:        50,
			handler:    anteHandlerStack,
			valid:      false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// msg and signatures
			fee := authtypes.NewStdFee(100000, sdk.NewCoins(sdk.NewInt64Coin("atom", tc.fee)))
			msgs := []sdk.Msg{sdk.NewTestMsg(tc.signer)}
			privs, accNums, seqs := []crypto.PrivKey{tc.signerKey}, []uint64{0}, []uint64{0}
			tx := authtypes.NewTestTxWithFeeAccount(ctx, msgs, privs, accNums, seqs, fee, tc.feeAccount)

			_, err := tc.handler(ctx, tx, false)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}

// don't cosume any gas
func SigGasNoConsumer(meter sdk.GasMeter, sig []byte, pubkey crypto.PubKey, params authtypes.Params) error {
	return nil
}
