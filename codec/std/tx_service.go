package std

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/tx/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client"
)

type TxServiceImpl struct{}

func (t TxServiceImpl) QueryTx(ctx context.Context, req *rest.QueryTxRequest) (*TxResponse, error) {
	res, err := client.QueryTx(rest.UnwrapCLIContext(ctx), req.TxHash)
	if err != nil {
		return nil, err
	}
	tx, ok := res.Tx.(Transaction)
	if !ok {
		return nil, fmt.Errorf("unable to decode transaction")
	}
	return &TxResponse{
		Base: &res.TxResponseBase,
		Tx:   &tx,
	}, nil
}

func (t TxServiceImpl) QueryTxs(context.Context, *rest.QueryTxsRequest) (*QueryTxsResponse, error) {
	panic("implement me")
}

func (t TxServiceImpl) BroadcastTx(ctx context.Context, req *BroadcastTxRequest) (*TxResponse, error) {
	//cliCtx := rest.UnwrapCLIContext(ctx)
	//switch req.Mode {
	//
	//}
	//cliCtx.BroadcastMode = req.Mode
	//res, err := cliCtx.BroadcastTx(txBytes)
	panic("implement me")
}

func (t TxServiceImpl) GenerateTx(context.Context, *GenerateTxRequest) (*Transaction, error) {
	panic("implement me")
}

var _ TxServiceServer = TxServiceImpl{}
