package testdata

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
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
