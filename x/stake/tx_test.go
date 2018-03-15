package stake

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	validator = []byte("addressvalidator1")
	empty     sdk.Address

	coinPos          = sdk.Coin{"fermion", 1000}
	coinZero         = sdk.Coin{"fermion", 0}
	coinNeg          = sdk.Coin{"fermion", -10000}
	coinPosNotAtoms  = sdk.Coin{"foo", 10000}
	coinZeroNotAtoms = sdk.Coin{"foo", 0}
	coinNegNotAtoms  = sdk.Coin{"foo", -10000}
)

func TestMsgAddrValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		address sdk.Address
		wantErr bool
	}{
		{"basic good", pks[0], false},
		{"empty delegator", crypto.PubKey{}, true},
	}

	for _, tc := range tests {
		tx := NewMsgAddr(tc.address)
		assert.Equal(t, tc.wantErr, tx.ValidateBasic() != nil,
			"test: %v, tx.ValidateBasic: %v", tc.name, tx.ValidateBasic())
	}
}

func TestValidateCoin(t *testing.T) {
	tests := []struct {
		name    string
		coin    sdk.Coin
		wantErr bool
	}{
		{"basic good", coinPos, false},
		{"zero coin", coinZero, true},
		{"neg coin", coinNeg, true},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.wantErr, tx.validateCoin(tc.coin) != nil,
			"test: %v, tx.ValidateBasic: %v", tc.name, tx.ValidateBasic())
	}
}

func TestAllAreTx(t *testing.T) {

	// make sure all types construct properly
	pubKey := newPubKey("1234567890")
	bondAmt := 1234321
	bond := sdk.Coin{Denom: "ATOM", Amount: int64(bondAmt)}

	txDelegate := NewMsgDelegate(bond, pubKey)
	_, ok := txDelegate.(MsgDelegate)
	assert.True(t, ok, "%#v", txDelegate)

	txUnbond := NewMsgUnbond(strconv.Itoa(bondAmt), pubKey)
	_, ok = txUnbond.(MsgUnbond)
	assert.True(t, ok, "%#v", txUnbond)

	txDecl := NewMsgDeclareCandidacy(bond, pubKey, Description{})
	_, ok = txDecl.(MsgDeclareCandidacy)
	assert.True(t, ok, "%#v", txDecl)

	txEditCan := NewMsgEditCandidacy(pubKey, Description{})
	_, ok = txEditCan.(MsgEditCandidacy)
	assert.True(t, ok, "%#v", txEditCan)
}

func TestSerializeTx(t *testing.T) {

	// make sure all types construct properly
	pubKey := newPubKey("1234567890")
	bondAmt := 1234321
	bond := sdk.Coin{Denom: "ATOM", Amount: int64(bondAmt)}

	tests := []struct {
		tx sdk.Tx
	}{
		{NewMsgUnbond(strconv.Itoa(bondAmt), pubKey)},
		{NewMsgDeclareCandidacy(bond, pubKey, Description{})},
		{NewMsgDeclareCandidacy(bond, pubKey, Description{})},
		// {NewTxRevokeCandidacy(pubKey)},
	}

	for i, tc := range tests {
		var tx sdk.Tx
		bs := wire.BinaryBytes(tc.tx)
		err := wire.ReadBinaryBytes(bs, &tx)
		if assert.NoError(t, err, "%d", i) {
			assert.Equal(t, tc.tx, tx, "%d", i)
		}
	}
}
