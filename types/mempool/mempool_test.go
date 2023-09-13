package mempool_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/log"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// testPubKey is a dummy implementation of PubKey used for testing.
type testPubKey struct {
	address sdk.AccAddress
}

func (t testPubKey) Reset() { panic("not implemented") }

func (t testPubKey) String() string { panic("not implemented") }

func (t testPubKey) ProtoMessage() { panic("not implemented") }

func (t testPubKey) Address() cryptotypes.Address { return t.address.Bytes() }

func (t testPubKey) Bytes() []byte { panic("not implemented") }

func (t testPubKey) VerifySignature(msg, sig []byte) bool { panic("not implemented") }

func (t testPubKey) Equals(key cryptotypes.PubKey) bool { panic("not implemented") }

func (t testPubKey) Type() string { panic("not implemented") }

// testTx is a dummy implementation of Tx used for testing.
type testTx struct {
	id       int
	priority int64
	nonce    uint64
	address  sdk.AccAddress
	// useful for debugging
	strAddress string
}

func (tx testTx) GetSigners() ([][]byte, error) { panic("not implemented") }

func (tx testTx) GetPubKeys() ([]cryptotypes.PubKey, error) { panic("not implemented") }

func (tx testTx) GetSignaturesV2() (res []txsigning.SignatureV2, err error) {
	res = append(res, txsigning.SignatureV2{
		PubKey:   testPubKey{address: tx.address},
		Data:     nil,
		Sequence: tx.nonce,
	})

	return res, nil
}

var (
	_ sdk.Tx                  = (*testTx)(nil)
	_ signing.SigVerifiableTx = (*testTx)(nil)
	_ cryptotypes.PubKey      = (*testPubKey)(nil)
)

func (tx testTx) GetMsgs() []sdk.Msg { return nil }

func (tx testTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }

func (tx testTx) ValidateBasic() error { return nil }

func (tx testTx) String() string {
	return fmt.Sprintf("tx a: %s, p: %d, n: %d", tx.address, tx.priority, tx.nonce)
}

type sigErrTx struct {
	getSigs func() ([]txsigning.SignatureV2, error)
}

func (sigErrTx) Size() int64 { return 0 }

func (sigErrTx) GetMsgs() []sdk.Msg { return nil }

func (sigErrTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }

func (sigErrTx) ValidateBasic() error { return nil }

func (sigErrTx) GetSigners() ([][]byte, error) { return nil, nil }

func (sigErrTx) GetPubKeys() ([]cryptotypes.PubKey, error) { return nil, nil }

func (t sigErrTx) GetSignaturesV2() ([]txsigning.SignatureV2, error) { return t.getSigs() }

type txSpec struct {
	i int
	p int
	n int
	a sdk.AccAddress
}

func (tx txSpec) String() string {
	return fmt.Sprintf("[tx i: %d, a: %s, p: %d, n: %d]", tx.i, tx.a, tx.p, tx.n)
}

func fetchTxs(iterator mempool.Iterator, maxBytes int64) []sdk.Tx {
	const txSize = 1
	var (
		txs      []sdk.Tx
		numBytes int64
	)
	for iterator != nil {
		if numBytes += txSize; numBytes > maxBytes {
			break
		}
		txs = append(txs, iterator.Tx())
		i := iterator.Next()
		iterator = i
	}
	return txs
}

func (s *MempoolTestSuite) TestDefaultMempool() {
	t := s.T()
	ctx := sdk.NewContext(nil, false, log.NewNopLogger())
	accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 10)
	txCount := 1000
	var txs []testTx

	for i := 0; i < txCount; i++ {
		acc := accounts[i%len(accounts)]
		tx := testTx{
			nonce:    0,
			address:  acc.Address,
			priority: rand.Int63(),
		}
		txs = append(txs, tx)
	}

	// empty mempool behavior
	require.Equal(t, 0, s.mempool.CountTx())
	itr := s.mempool.Select(ctx, nil)
	require.Nil(t, itr)

	// same sender-nonce just overwrites a tx
	for _, tx := range txs {
		ctx = ctx.WithPriority(tx.priority)
		err := s.mempool.Insert(ctx, tx)
		require.NoError(t, err)
	}
	require.Equal(t, len(accounts), s.mempool.CountTx())

	// distinct sender-nonce should not overwrite a tx
	s.resetMempool()
	for i, tx := range txs {
		tx.nonce = uint64(i)
		err := s.mempool.Insert(ctx, tx)
		require.NoError(t, err)
	}
	require.Equal(t, txCount, s.mempool.CountTx())

	itr = s.mempool.Select(ctx, nil)
	sel := fetchTxs(itr, 13)
	require.Equal(t, 13, len(sel))

	// a tx which does not implement SigVerifiableTx should not be inserted
	tx := &sigErrTx{getSigs: func() ([]txsigning.SignatureV2, error) {
		return nil, fmt.Errorf("error")
	}}
	require.Error(t, s.mempool.Insert(ctx, tx))
	require.Error(t, s.mempool.Remove(tx))
	tx.getSigs = func() ([]txsigning.SignatureV2, error) {
		return nil, nil
	}
	require.Error(t, s.mempool.Insert(ctx, tx))
	require.Error(t, s.mempool.Remove(tx))

	// removing a tx not in the mempool should error
	s.resetMempool()
	require.NoError(t, s.mempool.Insert(ctx, txs[0]))
	require.ErrorIs(t, s.mempool.Remove(txs[1]), mempool.ErrTxNotFound)

	// inserting a tx with a different priority should overwrite the old tx
	newPriorityTx := testTx{
		address:  txs[0].address,
		priority: txs[0].priority + 1,
		nonce:    txs[0].nonce,
	}
	require.NoError(t, s.mempool.Insert(ctx, newPriorityTx))
	require.Equal(t, 1, s.mempool.CountTx())
}

type MempoolTestSuite struct {
	suite.Suite
	numTxs      int
	numAccounts int
	iterations  int
	mempool     mempool.Mempool
}

func (s *MempoolTestSuite) resetMempool() {
	s.iterations = 0
	s.mempool = mempool.NewSenderNonceMempool()
}

func (s *MempoolTestSuite) SetupTest() {
	s.numTxs = 1000
	s.numAccounts = 100
	s.resetMempool()
}

func TestMempoolTestSuite(t *testing.T) {
	suite.Run(t, new(MempoolTestSuite))
}

func (s *MempoolTestSuite) TestSampleTxs() {
	ctxt := sdk.NewContext(nil, false, log.NewNopLogger())
	t := s.T()
	s.resetMempool()
	mp := s.mempool
	delegatorTx, err := unmarshalTx(msgWithdrawDelegatorReward)

	require.NoError(t, err)
	require.NoError(t, mp.Insert(ctxt, delegatorTx))
	require.Equal(t, 1, mp.CountTx())
}

func unmarshalTx(txBytes []byte) (sdk.Tx, error) {
	cfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{}, gov.AppModuleBasic{})
	return cfg.TxConfig.TxJSONDecoder()(txBytes)
}

var msgWithdrawDelegatorReward = []byte("{\"body\":{\"messages\":[{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1lzhlnpahvznwfv4jmay2tgaha5kmz5qxerarrl\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1k2d9ed9vgfuk2m58a2d80q9u6qljkh4vfaqjfq\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1vygmh344ldv9qefss9ek7ggsnxparljlmj56q5\"},{\"@type\":\"\\/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward\",\"delegator_address\":\"cosmos16w6g0whmw703t8h2m9qmq2fd9dwaw6fjszzjsw\",\"validator_address\":\"cosmosvaloper1ej2es5fjztqjcd4pwa0zyvaevtjd2y5wxxp9gd\"}],\"memo\":\"\",\"timeout_height\":\"0\",\"extension_options\":[],\"non_critical_extension_options\":[]},\"auth_info\":{\"signer_infos\":[{\"public_key\":{\"@type\":\"\\/cosmos.crypto.secp256k1.PubKey\",\"key\":\"AmbXAy10a0SerEefTYQzqyGQdX5kiTEWJZ1PZKX1oswX\"},\"mode_info\":{\"single\":{\"mode\":\"SIGN_MODE_LEGACY_AMINO_JSON\"}},\"sequence\":\"119\"}],\"fee\":{\"amount\":[{\"denom\":\"uatom\",\"amount\":\"15968\"}],\"gas_limit\":\"638717\",\"payer\":\"\",\"granter\":\"\"}},\"signatures\":[\"ji+inUo4xGlN9piRQLdLCeJWa7irwnqzrMVPcmzJyG5y6NPc+ZuNaIc3uvk5NLDJytRB8AHX0GqNETR\\/Q8fz4Q==\"]}")
