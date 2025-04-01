package mempool_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

type nonVerifiableTx struct{}

func (n nonVerifiableTx) GetMsgs() []sdk.Msg {
	panic("not implemented")
}

func (n nonVerifiableTx) GetMsgsV2() ([]proto.Message, error) {
	panic("not implemented")
}

func TestDefaultSignerExtractor(t *testing.T) {
	accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(0)), 1)
	sa := accounts[0].Address
	ext := mempool.NewDefaultSignerExtractionAdapter()
	goodTx := testTx{id: 0, priority: 0, nonce: 0, address: sa}
	badTx := &sigErrTx{getSigs: func() ([]txsigning.SignatureV2, error) {
		return nil, fmt.Errorf("error")
	}}
	nonSigVerify := nonVerifiableTx{}

	tests := []struct {
		name string
		tx   sdk.Tx
		sea  mempool.SignerExtractionAdapter
		err  error
	}{
		{name: "valid tx extracts sigs", tx: goodTx, sea: ext, err: nil},
		{name: "invalid tx fails on sig", tx: badTx, sea: ext, err: fmt.Errorf("err")},
		{name: "non-verifiable tx fails on conversion", tx: nonSigVerify, sea: ext, err: fmt.Errorf("tx of type %T does not implement SigVerifiableTx", nonSigVerify)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sigs, err := test.sea.GetSigners(test.tx)
			if test.err != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, sigs[0].String(), mempool.SignerData{Signer: sa, Sequence: 0}.String())
		})
	}
}
