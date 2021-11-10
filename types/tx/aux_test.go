package tx_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func TestSignDocDirectAux(t *testing.T) {
	bodyBz := []byte{42}
	_, pk, addr := testdata.KeyTestPubAddr()
	pkAny, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)

	testcases := []struct {
		name   string
		sd     tx.SignDocDirectAux
		expErr bool
	}{
		{"empty bodyBz", tx.SignDocDirectAux{}, true},
		{"empty pubkey", tx.SignDocDirectAux{BodyBytes: bodyBz}, true},
		{"empty tip amount", tx.SignDocDirectAux{BodyBytes: bodyBz, PublicKey: pkAny, Tip: &tx.Tip{Tipper: addr.String()}}, true},
		{"empty tipper", tx.SignDocDirectAux{BodyBytes: bodyBz, PublicKey: pkAny, Tip: &tx.Tip{Amount: testdata.NewTestFeeAmount()}}, true},
		{"happy case w/o tip", tx.SignDocDirectAux{BodyBytes: bodyBz, PublicKey: pkAny}, false},
		{"happy case w/ tip", tx.SignDocDirectAux{
			BodyBytes: bodyBz,
			PublicKey: pkAny,
			Tip:       &tx.Tip{Tipper: addr.String(), Amount: testdata.NewTestFeeAmount()},
		}, false},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sd.ValidateBasic()

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuxSignerData(t *testing.T) {
	bodyBz := []byte{42}
	_, pk, addr := testdata.KeyTestPubAddr()
	pkAny, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)
	sig := []byte{42}
	sd := &tx.SignDocDirectAux{BodyBytes: bodyBz, PublicKey: pkAny}

	testcases := []struct {
		name   string
		sd     tx.AuxSignerData
		expErr bool
	}{
		{"empty address", tx.AuxSignerData{}, true},
		{"empty sign mode", tx.AuxSignerData{Address: addr.String()}, true},
		{"SIGN_MODE_DIRECT", tx.AuxSignerData{Address: addr.String(), Mode: signing.SignMode(signing.SignMode_SIGN_MODE_DIRECT)}, true},
		{"no sig", tx.AuxSignerData{Address: addr.String(), Mode: signing.SignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX)}, true},
		{"happy case WITH DIRECT_AUX", tx.AuxSignerData{Address: addr.String(), Mode: signing.SignMode_SIGN_MODE_DIRECT_AUX, SignDoc: sd, Sig: sig}, false},
		{"happy case WITH DIRECT_AUX", tx.AuxSignerData{Address: addr.String(), Mode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, SignDoc: sd, Sig: sig}, false},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sd.ValidateBasic()

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
