package auth

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

// msg type for testing
type testMsg struct {
	signers []sdk.Address
}

func newTestMsg(addrs ...sdk.Address) *testMsg {
	return &testMsg{
		signers: addrs,
	}
}

func (msg *testMsg) Type() string                            { return "testMsg" }
func (msg *testMsg) Get(key interface{}) (value interface{}) { return nil }
func (msg *testMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg.signers)
	if err != nil {
		panic(err)
	}
	return bz
}
func (msg *testMsg) ValidateBasic() sdk.Error { return nil }
func (msg *testMsg) GetSigners() []sdk.Address {
	return msg.signers
}

func newStdFee() sdk.StdFee {
	return sdk.NewStdFee(100,
		sdk.Coin{"atom", 150},
	)
}

// coins to more than cover the fee
func newCoins() sdk.Coins {
	return sdk.Coins{
		{"atom", 10000000},
	}
}

// generate a priv key and return it with its address
func privAndAddr() (crypto.PrivKey, sdk.Address) {
	priv := crypto.GenPrivKeyEd25519()
	addr := priv.PubKey().Address()
	return priv.Wrap(), addr
}

// run the tx through the anteHandler and ensure its valid
func checkValidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx) {
	_, result, abort := anteHandler(ctx, tx)
	assert.False(t, abort)
	assert.Equal(t, sdk.CodeOK, result.Code)
	assert.True(t, result.IsOK())
}

// run the tx through the anteHandler and ensure it fails with the given code
func checkInvalidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, code sdk.CodeType) {
	_, result, abort := anteHandler(ctx, tx)
	assert.True(t, abort)
	assert.Equal(t, code, result.Code)
}

func newTestTx(ctx sdk.Context, msg sdk.Msg, privs []crypto.PrivKey, seqs []int64, fee sdk.StdFee) sdk.Tx {
	signBytes := sdk.StdSignBytes(ctx.ChainID(), seqs, fee, msg)
	return newTestTxWithSignBytes(msg, privs, seqs, fee, signBytes)
}

func newTestTxWithSignBytes(msg sdk.Msg, privs []crypto.PrivKey, seqs []int64, fee sdk.StdFee, signBytes []byte) sdk.Tx {
	sigs := make([]sdk.StdSignature, len(privs))
	for i, priv := range privs {
		sigs[i] = sdk.StdSignature{PubKey: priv.PubKey(), Signature: priv.Sign(signBytes), Sequence: seqs[i]}
	}
	tx := sdk.NewStdTx(msg, fee, sigs)
	return tx
}

// Test various error cases in the AnteHandler control flow.
func TestAnteHandlerSigErrors(t *testing.T) {
	// setup
	ms, capKey := setupMultiStore()
	mapper := NewAccountMapper(capKey, &BaseAccount{})
	anteHandler := NewAnteHandler(mapper)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "mychainid"}, false, nil)

	// keys and addresses
	priv1, addr1 := privAndAddr()
	priv2, addr2 := privAndAddr()

	// msg and signatures
	var tx sdk.Tx
	msg := newTestMsg(addr1, addr2)
	fee := newStdFee()

	// test no signatures
	privs, seqs := []crypto.PrivKey{}, []int64{}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeUnauthorized)

	// test num sigs dont match GetSigners
	privs, seqs = []crypto.PrivKey{priv1}, []int64{0}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeUnauthorized)

	// test an unrecognized account
	privs, seqs = []crypto.PrivKey{priv1, priv2}, []int64{0, 0}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeUnrecognizedAddress)

	// save the first account, but second is still unrecognized
	acc1 := mapper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(fee.Amount)
	mapper.SetAccount(ctx, acc1)
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeUnrecognizedAddress)
}

// Test logic around sequence checking with one signer and many signers.
func TestAnteHandlerSequences(t *testing.T) {
	// setup
	ms, capKey := setupMultiStore()
	mapper := NewAccountMapper(capKey, &BaseAccount{})
	anteHandler := NewAnteHandler(mapper)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "mychainid"}, false, nil)

	// keys and addresses
	priv1, addr1 := privAndAddr()
	priv2, addr2 := privAndAddr()

	// set the accounts
	acc1 := mapper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	mapper.SetAccount(ctx, acc1)
	acc2 := mapper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	mapper.SetAccount(ctx, acc2)

	// msg and signatures
	var tx sdk.Tx
	msg := newTestMsg(addr1)
	fee := newStdFee()

	// test good tx from one signer
	privs, seqs := []crypto.PrivKey{priv1}, []int64{0}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx)

	// test sending it again fails (replay protection)
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeInvalidSequence)

	// fix sequence, should pass
	seqs = []int64{1}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx)

	// new tx with another signer and correct sequences
	msg = newTestMsg(addr1, addr2)
	privs, seqs = []crypto.PrivKey{priv1, priv2}, []int64{2, 0}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx)

	// replay fails
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeInvalidSequence)

	// tx from just second signer with incorrect sequence fails
	msg = newTestMsg(addr2)
	privs, seqs = []crypto.PrivKey{priv2}, []int64{0}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeInvalidSequence)

	// fix the sequence and it passes
	tx = newTestTx(ctx, msg, []crypto.PrivKey{priv2}, []int64{1}, fee)
	checkValidTx(t, anteHandler, ctx, tx)

	// another tx from both of them that passes
	msg = newTestMsg(addr1, addr2)
	privs, seqs = []crypto.PrivKey{priv1, priv2}, []int64{3, 2}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx)
}

// Test logic around fee deduction.
func TestAnteHandlerFees(t *testing.T) {
	// // setup
	// ms, capKey := setupMultiStore()
	// mapper := NewAccountMapper(capKey, &BaseAccount{})
	// anteHandler := NewAnteHandler(mapper)
	// ctx := sdk.NewContext(ms, abci.Header{ChainID: "mychainid"}, false, nil)
	//
	// // keys and addresses
	// priv1, addr1 := privAndAddr()
	// priv2, addr2 := privAndAddr()
	//
	// // set the accounts
	// acc1 := mapper.NewAccountWithAddress(ctx, addr1)
	// mapper.SetAccount(ctx, acc1)
	// acc2 := mapper.NewAccountWithAddress(ctx, addr2)
	// mapper.SetAccount(ctx, acc2)
	//
	// // msg and signatures
	// var tx sdk.Tx
	// msg := newTestMsg(addr1)
	// tx = newTestTx(ctx, msg, []crypto.PrivKey{priv1}, []int64{0}, int64(1))

	// TODO
}

func TestAnteHandlerBadSignBytes(t *testing.T) {
	// setup
	ms, capKey := setupMultiStore()
	mapper := NewAccountMapper(capKey, &BaseAccount{})
	anteHandler := NewAnteHandler(mapper)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "mychainid"}, false, nil)

	// keys and addresses
	priv1, addr1 := privAndAddr()
	priv2, addr2 := privAndAddr()

	// set the accounts
	acc1 := mapper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	mapper.SetAccount(ctx, acc1)
	acc2 := mapper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	mapper.SetAccount(ctx, acc2)

	var tx sdk.Tx
	msg := newTestMsg(addr1)
	fee := newStdFee()

	// test good tx and signBytes
	privs, seqs := []crypto.PrivKey{priv1}, []int64{0}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx)

	chainID := ctx.ChainID()
	codeUnauth := sdk.CodeUnauthorized

	cases := []struct {
		chainID string
		seqs    []int64
		fee     sdk.StdFee
		msg     sdk.Msg
		code    sdk.CodeType
	}{
		{"", []int64{1}, fee, msg, codeUnauth},                    // test invalid chain_id
		{chainID, []int64{2}, fee, msg, codeUnauth},               // test wrong seqs
		{chainID, []int64{1}, fee, newTestMsg(addr2), codeUnauth}, // test wrong msg
	}

	privs, seqs = []crypto.PrivKey{priv1}, []int64{1}
	for _, cs := range cases {
		tx := newTestTxWithSignBytes(
			msg, privs, seqs, fee,
			sdk.StdSignBytes(cs.chainID, cs.seqs, cs.fee, cs.msg),
		)
		checkInvalidTx(t, anteHandler, ctx, tx, cs.code)
	}

	// test wrong signer if public key exist
	privs, seqs = []crypto.PrivKey{priv2}, []int64{1}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeUnauthorized)

	// test wrong signer if public doesn't exist
	msg = newTestMsg(addr2)
	privs, seqs = []crypto.PrivKey{priv1}, []int64{0}
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeInvalidPubKey)

}

func TestAnteHandlerSetPubKey(t *testing.T) {
	// setup
	ms, capKey := setupMultiStore()
	mapper := NewAccountMapper(capKey, &BaseAccount{})
	anteHandler := NewAnteHandler(mapper)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "mychainid"}, false, nil)

	// keys and addresses
	priv1, addr1 := privAndAddr()
	_, addr2 := privAndAddr()

	// set the accounts
	acc1 := mapper.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(newCoins())
	mapper.SetAccount(ctx, acc1)
	acc2 := mapper.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(newCoins())
	mapper.SetAccount(ctx, acc2)

	var tx sdk.Tx

	// test good tx and set public key
	msg := newTestMsg(addr1)
	privs, seqs := []crypto.PrivKey{priv1}, []int64{0}
	fee := newStdFee()
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx)

	acc1 = mapper.GetAccount(ctx, addr1)
	require.Equal(t, acc1.GetPubKey(), priv1.PubKey())

	// test public key not found
	msg = newTestMsg(addr2)
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	sigs := tx.GetSignatures()
	sigs[0].PubKey = crypto.PubKey{}
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeInvalidPubKey)

	acc2 = mapper.GetAccount(ctx, addr2)
	assert.True(t, acc2.GetPubKey().Empty())

	// test invalid signature and public key
	tx = newTestTx(ctx, msg, privs, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, sdk.CodeInvalidPubKey)

	acc2 = mapper.GetAccount(ctx, addr2)
	assert.True(t, acc2.GetPubKey().Empty())
}
