package ante_test

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// TestDeductFeeDecorator_FeegrantErrorIsBech32 ensures the error returned when a
// fee grant does not allow paying the fee contains the bech32 addresses rather
// than raw address bytes. Formatting the raw []byte addresses with %s produced
// control characters in the error (and the tx log), which can break UTF-8/YAML
// rendering on the client side.
func TestDeductFeeDecorator_FeegrantErrorIsBech32(t *testing.T) {
	s := SetupTestSuite(t, false)
	protoTxCfg := tx.NewTxConfig(codec.NewProtoCodec(s.encCfg.InterfaceRegistry), tx.DefaultSignModes)
	dfd := ante.NewDeductFeeDecorator(s.accountKeeper, s.bankKeeper, s.feeGrantKeeper, nil)
	feeAnteHandler := sdk.ChainAnteDecorators(dfd)

	accs := s.CreateTestAccounts(2)
	payer := accs[0]
	granter := accs[1].acc.GetAddress()

	s.feeGrantKeeper.EXPECT().
		UseGrantedFees(gomock.Any(), granter, payer.acc.GetAddress(), gomock.Any(), gomock.Any()).
		Return(sdkerrors.ErrNotFound.Wrap("fee-grant not found")).
		AnyTimes()

	fee := sdk.NewCoins(sdk.NewInt64Coin("atom", 2))
	msgs := []sdk.Msg{testdata.NewTestMsg(payer.acc.GetAddress())}
	acc := s.accountKeeper.GetAccount(s.ctx, payer.acc.GetAddress())
	tx, err := genTxWithFeeGranter(protoTxCfg, msgs, fee, 10_000_000, s.ctx.ChainID(),
		[]uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, granter,
		[]cryptotypes.PrivKey{payer.priv}...)
	require.NoError(t, err)

	_, err = feeAnteHandler(s.ctx, tx, false)
	require.Error(t, err)
	msg := err.Error()

	require.True(t, utf8.ValidString(msg), "error must be valid UTF-8, got %q", msg)
	for _, r := range msg {
		if r != '\n' && r != '\t' {
			require.Falsef(t, unicode.IsControl(r), "error contains control char %U: %q", r, msg)
		}
	}
	require.Contains(t, msg, granter.String())
	require.Contains(t, msg, payer.acc.GetAddress().String())
	require.True(t, strings.Contains(msg, "does not allow to pay fees for"))
}
