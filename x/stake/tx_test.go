package stake

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/coin"

	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
)

var (
	validator = sdk.Actor{"testChain", "testapp", []byte("addressvalidator1")}
	empty     sdk.Actor

	coinPos          = coin.Coin{"fermion", 1000}
	coinZero         = coin.Coin{"fermion", 0}
	coinNeg          = coin.Coin{"fermion", -10000}
	coinPosNotAtoms  = coin.Coin{"foo", 10000}
	coinZeroNotAtoms = coin.Coin{"foo", 0}
	coinNegNotAtoms  = coin.Coin{"foo", -10000}
)

func TestBondUpdateValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		PubKey  crypto.PubKey
		Bond    coin.Coin
		wantErr bool
	}{
		{"basic good", pks[0], coinPos, false},
		{"empty delegator", crypto.PubKey{}, coinPos, true},
		{"zero coin", pks[0], coinZero, true},
		{"neg coin", pks[0], coinNeg, true},
	}

	for _, tc := range tests {
		tx := TxDelegate{BondUpdate{
			PubKey: tc.PubKey,
			Bond:   tc.Bond,
		}}
		assert.Equal(t, tc.wantErr, tx.ValidateBasic() != nil,
			"test: %v, tx.ValidateBasic: %v", tc.name, tx.ValidateBasic())
	}
}

func TestAllAreTx(t *testing.T) {
	assert := assert.New(t)

	// make sure all types construct properly
	pubKey := newPubKey("1234567890")
	bondAmt := 1234321
	bond := coin.Coin{Denom: "ATOM", Amount: int64(bondAmt)}

	// Note that Wrap is only defined on BondUpdate, so when you call it,
	// you lose all info on the embedding type. Please add Wrap()
	// method to all the parents
	txDelegate := NewTxDelegate(bond, pubKey)
	_, ok := txDelegate.Unwrap().(TxDelegate)
	assert.True(ok, "%#v", txDelegate)

	txUnbond := NewTxUnbond(strconv.Itoa(bondAmt), pubKey)
	_, ok = txUnbond.Unwrap().(TxUnbond)
	assert.True(ok, "%#v", txUnbond)

	txDecl := NewTxDeclareCandidacy(bond, pubKey, Description{})
	_, ok = txDecl.Unwrap().(TxDeclareCandidacy)
	assert.True(ok, "%#v", txDecl)

	txEditCan := NewTxEditCandidacy(pubKey, Description{})
	_, ok = txEditCan.Unwrap().(TxEditCandidacy)
	assert.True(ok, "%#v", txEditCan)
}

func TestSerializeTx(t *testing.T) {
	assert := assert.New(t)

	// make sure all types construct properly
	pubKey := newPubKey("1234567890")
	bondAmt := 1234321
	bond := coin.Coin{Denom: "ATOM", Amount: int64(bondAmt)}

	tests := []struct {
		tx sdk.Tx
	}{
		{NewTxUnbond(strconv.Itoa(bondAmt), pubKey)},
		{NewTxDeclareCandidacy(bond, pubKey, Description{})},
		{NewTxDeclareCandidacy(bond, pubKey, Description{})},
		// {NewTxRevokeCandidacy(pubKey)},
	}

	for i, tc := range tests {
		var tx sdk.Tx
		bs := wire.BinaryBytes(tc.tx)
		err := wire.ReadBinaryBytes(bs, &tx)
		if assert.NoError(err, "%d", i) {
			assert.Equal(tc.tx, tx, "%d", i)
		}
	}
}
