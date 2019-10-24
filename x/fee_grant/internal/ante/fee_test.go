package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/internal/ante"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/internal/types"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/internal/types/tx"
)

// newAnteHandler is just like auth.NewAnteHandler, except we use the DeductGrantedFeeDecorator
// in order to allow payment of fees via a grant.
//
// This is used for our full-stack tests
func newAnteHandler(ak authkeeper.AccountKeeper, supplyKeeper authtypes.SupplyKeeper, dk keeper.Keeper, sigGasConsumer authante.SignatureVerificationGasConsumer) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		authante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		authante.NewMempoolFeeDecorator(),
		authante.NewValidateBasicDecorator(),
		authante.NewValidateMemoDecorator(ak),
		authante.NewConsumeGasForTxSizeDecorator(ak),
		// DeductGrantedFeeDecorator will create an empty account if we sign with no tokens but valid validation
		// This must be before SetPubKey, ValidateSigCount, SigVerification, which error if account doesn't exist yet
		ante.NewDeductGrantedFeeDecorator(ak, supplyKeeper, dk),
		authante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		authante.NewValidateSigCountDecorator(ak),
		authante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		authante.NewSigVerificationDecorator(ak),
		authante.NewIncrementSequenceDecorator(ak), // innermost AnteDecorator
	)
}

func TestDeductFeesNoDelegation(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)

	// this just tests our handler
	dfd := ante.NewDeductGrantedFeeDecorator(app.AccountKeeper, app.SupplyKeeper, app.FeeGrantKeeper)
	ourAnteHandler := sdk.ChainAnteDecorators(dfd)

	// this tests the whole stack
	anteHandlerStack := newAnteHandler(app.AccountKeeper, app.SupplyKeeper, app.FeeGrantKeeper, SigGasNoConsumer)

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

	// Set grant from addr2 to addr3 (plenty to pay)
	app.FeeGrantKeeper.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr2,
		Grantee: addr3,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 500)),
		},
	})

	// Set low grant from addr2 to addr4 (keeper will reject)
	app.FeeGrantKeeper.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr2,
		Grantee: addr4,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 20)),
		},
	})

	// Set grant from addr1 to addr4 (cannot cover this )
	app.FeeGrantKeeper.GrantFeeAllowance(ctx, types.FeeAllowanceGrant{
		Granter: addr2,
		Grantee: addr3,
		Allowance: &types.BasicFeeAllowance{
			SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("atom", 500)),
		},
	})

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
		"valid fee grant without account (only ours)": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr2,
			fee:        50,
			handler:    ourAnteHandler,
			valid:      true,
		},
		"no fee grant (only ours)": {
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
		"granter cannot cover allowed fee grant (only ours)": {
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
		"valid fee grant without account (whole stack)": {
			signerKey:  priv3,
			signer:     addr3,
			feeAccount: addr2,
			fee:        50,
			handler:    anteHandlerStack,
			valid:      true,
		},
		"no fee grant (whole stack)": {
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
		"granter cannot cover allowed fee grant (whole stack)": {
			signerKey:  priv4,
			signer:     addr4,
			feeAccount: addr1,
			fee:        50,
			handler:    anteHandlerStack,
			valid:      false,
		},
	}

	for name, stc := range cases {
		tc := stc // to make scopelint happy
		t.Run(name, func(t *testing.T) {
			// msg and signatures
			fee := tx.NewGrantedFee(100000, sdk.NewCoins(sdk.NewInt64Coin("atom", tc.fee)), tc.feeAccount)
			msgs := []sdk.Msg{sdk.NewTestMsg(tc.signer)}
			privs, accNums, seqs := []crypto.PrivKey{tc.signerKey}, []uint64{0}, []uint64{0}

			tx := tx.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

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
