package accounts

import (
	"context"
	"fmt"

	v1 "cosmossdk.io/x/accounts/v1"
)

var _ v1.QueryServer = queryServer{}

func NewQueryServer(k Keeper) v1.QueryServer {
	return &queryServer{k}
}

type queryServer struct {
	k Keeper
}

func (q queryServer) AccountQuery(ctx context.Context, request *v1.AccountQueryRequest) (*v1.AccountQueryResponse, error) {
	// get target addr
	targetAddr, err := q.k.addressCodec.StringToBytes(request.Target)
	if err != nil {
		return nil, err
	}

	// get acc type
	accType, err := q.k.AccountsByType.Get(ctx, targetAddr)
	if err != nil {
		return nil, err
	}

	// get impl
	impl, err := q.k.getImplementation(accType)
	if err != nil {
		return nil, err
	}

	// decode req into boxed concrete type
	queryReq, err := impl.DecodeQueryRequest(request.Request)
	if err != nil {
		return nil, err
	}
	// run query
	resp, err := q.k.Query(ctx, targetAddr, queryReq)
	if err != nil {
		return nil, err
	}

	// encode response
	respBytes, err := impl.EncodeQueryResponse(resp)
	if err != nil {
		return nil, err
	}

	return &v1.AccountQueryResponse{
		Response: respBytes,
	}, nil
}

func (q queryServer) Schema(_ context.Context, request *v1.SchemaRequest) (*v1.SchemaResponse, error) {
	// TODO: maybe we should cache this, considering accounts types are not
	// added on the fly as the chain is running.
	schemas := v1.MakeAccountsSchemas(q.k.accounts)
	schema, ok := schemas[request.AccountType]
	if !ok {
		return nil, fmt.Errorf("%w: %s", errAccountTypeNotFound, request.AccountType)
	}
	return schema, nil
}

func (q queryServer) AccountType(ctx context.Context, request *v1.AccountTypeRequest) (*v1.AccountTypeResponse, error) {
	addr, err := q.k.addressCodec.StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}
	accType, err := q.k.AccountsByType.Get(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &v1.AccountTypeResponse{
		AccountType: accType,
	}, nil
}
