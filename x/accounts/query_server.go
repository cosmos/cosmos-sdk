package accounts

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"cosmossdk.io/x/accounts/internal/implementation"
	v1 "cosmossdk.io/x/accounts/v1"
)

var _ v1.QueryServer = &queryServer{}

// NewQueryServer initializes a new instance of QueryServer.
// It precalculates and stores schemas for efficient schema retrieval.
func NewQueryServer(k Keeper) v1.QueryServer {
	return &queryServer{
		k: k,
	}
}

type queryServer struct {
	k Keeper
}

func (q *queryServer) AccountQuery(ctx context.Context, request *v1.AccountQueryRequest) (*v1.AccountQueryResponse, error) {
	// get target addr
	targetAddr, err := q.k.addressCodec.StringToBytes(request.Target)
	if err != nil {
		return nil, err
	}

	// decode req into boxed concrete type
	queryReq, err := implementation.UnpackAnyRaw(request.Request)
	if err != nil {
		return nil, err
	}
	// run query
	resp, err := q.k.Query(ctx, targetAddr, queryReq)
	if err != nil {
		return nil, err
	}

	// encode response
	respAny, err := implementation.PackAny(resp)
	if err != nil {
		return nil, err
	}

	return &v1.AccountQueryResponse{
		Response: respAny,
	}, nil
}

// Schema retrieves the schema for a given account.
// It checks the precalculated schemas and returns an error if the schema is not found.
func (q *queryServer) Schema(ctx context.Context, request *v1.SchemaRequest) (*v1.SchemaResponse, error) {
	// get account
	addr, err := q.k.addressCodec.StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}

	impl, err := q.k.getImplementation(ctx, addr)
	if err != nil {
		return nil, err
	}
	return makeAccountSchema(ctx, impl)
}

func (q *queryServer) InitSchema(ctx context.Context, request *v1.InitSchemaRequest) (*v1.InitSchemaResponse, error) {
	accTypes, exists := q.k.accounts[request.AccountType]
	if !exists {
		return nil, fmt.Errorf("account type %s not found", request.AccountType)
	}

	initSchema, err := accTypes.GetInitHandlerSchema(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.InitSchemaResponse{
		InitSchema: &v1.Handler{
			Request:  initSchema.RequestSchema.Name,
			Response: initSchema.ResponseSchema.Name,
		},
	}, nil
}

func makeAccountSchema(
	ctx context.Context,
	impl implementation.Implementation,
) (*v1.SchemaResponse, error) {
	initSchema, err := impl.GetInitHandlerSchema(ctx)
	if err != nil {
		return nil, err
	}
	executeHandlers, err := impl.GetExecuteHandlersSchema(ctx)
	if err != nil {
		return nil, err
	}

	queryHandlers, err := impl.GetQueryHandlersSchema(ctx)
	if err != nil {
		return nil, err
	}

	return &v1.SchemaResponse{
		InitSchema: &v1.Handler{
			Request:  initSchema.RequestSchema.Name,
			Response: initSchema.ResponseSchema.Name,
		},
		ExecuteHandlers: makeHandlersSchema(executeHandlers),
		QueryHandlers:   makeHandlersSchema(queryHandlers),
	}, nil
}

func makeHandlersSchema(handlers map[string]implementation.HandlerSchema) []*v1.Handler {
	schemas := make([]*v1.Handler, 0, len(handlers))
	for name, handler := range handlers {
		schemas = append(schemas, &v1.Handler{
			Request:  name,
			Response: handler.ResponseSchema.Name,
		})
	}
	slices.SortFunc(schemas, func(a, b *v1.Handler) int {
		return strings.Compare(a.Request, b.Request)
	})
	return schemas
}

func (q *queryServer) AccountType(ctx context.Context, request *v1.AccountTypeRequest) (*v1.AccountTypeResponse, error) {
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

func (q *queryServer) AccountNumber(ctx context.Context, request *v1.AccountNumberRequest) (*v1.AccountNumberResponse, error) {
	addr, err := q.k.addressCodec.StringToBytes(request.Address)
	if err != nil {
		return nil, err
	}
	number, err := q.k.AccountByNumber.Get(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &v1.AccountNumberResponse{Number: number}, nil
}

const (
	// TODO(tip): evaluate if the following numbers should be parametrised over state, or over the node.
	SimulateAuthenticateGasLimit   = 1_000_000
	SimulateBundlerPaymentGasLimit = SimulateAuthenticateGasLimit
	ExecuteGasLimit                = SimulateAuthenticateGasLimit
)
