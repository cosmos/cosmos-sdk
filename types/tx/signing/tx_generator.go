package signing

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
)

type TxGenerator struct {
	Marshaler codec.Marshaler
}

var _ context.TxGenerator = TxGenerator{}

func (t TxGenerator) NewTx() context.TxBuilder {
	return TxBuilder{}
}

func (t TxGenerator) NewFee() context.ClientFee {
	panic("implement me")
}

func (t TxGenerator) NewSignature() context.ClientSignature {
	panic("implement me")
}

func (t TxGenerator) MarshalTx(tx sdk.Tx) ([]byte, error) {
	ptx, ok := tx.(*types.Tx)
	if !ok {
		return nil, fmt.Errorf("expected protobuf Tx, got %T", tx)
	}
	return t.Marshaler.MarshalBinaryBare(ptx)
}

type TxBuilder struct {
	*types.Tx
}

var _ context.TxBuilder = TxBuilder{}

func (t TxBuilder) GetTx() sdk.Tx {
	return t.Tx
}

func (t TxBuilder) SetMsgs(msgs ...sdk.Msg) error {
	anys := make([]*codectypes.Any, len(msgs))
	for i, msg := range msgs {
		pmsg, ok := msg.(proto.Message)
		if !ok {
			return fmt.Errorf("cannot proto marshal %T", msg)
		}
		any, err := codectypes.NewAnyWithValue(pmsg)
		if err != nil {
			return err
		}
		anys[i] = any
	}
	t.Body.Messages = anys
	return nil
}

func (t TxBuilder) GetSignatures() []sdk.Signature {
	panic("implement me")
}

func (t TxBuilder) SetSignatures(signature ...context.ClientSignature) error {
	panic("implement me")
}

func (t TxBuilder) GetFee() sdk.Fee {
	panic("implement me")
}

func (t TxBuilder) SetFee(fee context.ClientFee) error {
	panic("implement me")
}

func (t TxBuilder) GetMemo() string {
	panic("implement me")
}

func (t TxBuilder) SetMemo(s string) {
	panic("implement me")
}

func (t TxBuilder) CanonicalSignBytes(cid string, num, seq uint64) ([]byte, error) {
	panic("implement me")
}
