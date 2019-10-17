package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/delegation/internal/ante"
	// delTypes "github.com/cosmos/cosmos-sdk/x/delegation/internal/types"
)

func TestDeductFeesNoDelegation(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)
	dfd := ante.NewDeductDelegatedFeeDecorator(app.AccountKeeper, app.SupplyKeeper, app.DelegationKeeper)
	antehandler := sdk.ChainAnteDecorators(dfd)

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

	// Set delegation from addr1 to addr4 (insufficient funds)

	cases := map[string]struct {
		signerKey  crypto.PrivKey
		signer     sdk.AccAddress
		feeAccount sdk.AccAddress
		fee        int64
		valid      bool
	}{
		"paying with low funds": {
			signerKey: priv1,
			signer:    addr1,
			fee:       50,
			valid:     false,
		},
		"paying with good funds": {
			signerKey: priv2,
			signer:    addr2,
			fee:       50,
			valid:     true,
		},
		"paying with no account": {
			signerKey: priv3,
			signer:    addr3,
			fee:       1,
			valid:     false,
		},
		"no fee with real account": {
			signerKey: priv1,
			signer:    addr1,
			fee:       0,
			valid:     true,
		},
		"no fee with no account": {
			signerKey: priv4,
			signer:    addr4,
			fee:       0,
			valid:     true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// msg and signatures
			fee := authtypes.NewStdFee(100000, sdk.NewCoins(sdk.NewInt64Coin("atom", tc.fee)))
			msgs := []sdk.Msg{sdk.NewTestMsg(tc.signer)}
			privs, accNums, seqs := []crypto.PrivKey{tc.signerKey}, []uint64{0}, []uint64{0}
			tx := authtypes.NewTestTxWithFeeAccount(ctx, msgs, privs, accNums, seqs, fee, tc.feeAccount)

			_, err := antehandler(ctx, tx, false)
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
