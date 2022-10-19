package testdata

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	grpc "google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// iterCount defines the number of iterations to run on each query to test
	// determinism.
	iterCount = 1000
)

type QueryImpl struct{}

var _ QueryServer = QueryImpl{}

func (e QueryImpl) TestAny(_ context.Context, request *TestAnyRequest) (*TestAnyResponse, error) {
	animal, ok := request.AnyAnimal.GetCachedValue().(Animal)
	if !ok {
		return nil, fmt.Errorf("expected Animal")
	}

	any, err := types.NewAnyWithValue(animal.(proto.Message))
	if err != nil {
		return nil, err
	}

	return &TestAnyResponse{HasAnimal: &HasAnimal{
		Animal: any,
		X:      10,
	}}, nil
}

func (e QueryImpl) Echo(_ context.Context, req *EchoRequest) (*EchoResponse, error) {
	return &EchoResponse{Message: req.Message}, nil
}

func (e QueryImpl) SayHello(_ context.Context, request *SayHelloRequest) (*SayHelloResponse, error) {
	greeting := fmt.Sprintf("Hello %s!", request.Name)
	return &SayHelloResponse{Greeting: greeting}, nil
}

var _ types.UnpackInterfacesMessage = &TestAnyRequest{}

func (m *TestAnyRequest) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var animal Animal
	return unpacker.UnpackAny(m.AnyAnimal, &animal)
}

var _ types.UnpackInterfacesMessage = &TestAnyResponse{}

func (m *TestAnyResponse) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return m.HasAnimal.UnpackInterfaces(unpacker)
}

// QueryReq is a generic func to handle deterministic query requests.
func QueryReq[request proto.Message, response proto.Message](
	ctx sdk.Context,
	require *require.Assertions,
	req request,
	grpcFn func(context.Context, request, ...grpc.CallOption) (response, error),
	gasConsumed uint64,
	gasOverwrite bool,
) {
	before := ctx.GasMeter().GasConsumed()
	prevRes, err := grpcFn(ctx, req)
	require.NoError(err)
	if gasOverwrite { // to handle regressions
		gasConsumed = ctx.GasMeter().GasConsumed() - before
	}

	for i := 0; i < iterCount; i++ {
		before := ctx.GasMeter().GasConsumed()
		res, err := grpcFn(ctx, req)
		require.Equal(ctx.GasMeter().GasConsumed()-before, gasConsumed)
		require.NoError(err)
		require.Equal(res, prevRes)
	}
}
