package reflection_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/simapp"
)

type IntegrationTestSuite struct {
	suite.Suite

	queryClient reflection.ReflectionServiceClient
}

func (s *IntegrationTestSuite) SetupSuite() {
	app := simapp.Setup(false)

	sdkCtx := app.BaseApp.NewContext(false, tmproto.Header{})
	queryHelper := baseapp.NewQueryServerTestHelper(sdkCtx, app.InterfaceRegistry())
	queryClient := reflection.NewReflectionServiceClient(queryHelper)
	s.queryClient = queryClient
}

func (s IntegrationTestSuite) TestSimulateService() {
	// We will test the following interface for testing.
	var iface = "cosmos.evidence.v1beta1.Evidence"

	// Test that "cosmos.evidence.v1beta1.Evidence" is included in the
	// interfaces.
	resIface, err := s.queryClient.ListAllInterfaces(
		context.Background(),
		&reflection.ListAllInterfacesRequest{},
	)
	s.Require().NoError(err)
	s.Require().Contains(resIface.GetInterfaceNames(), iface)

	// Test that "cosmos.evidence.v1beta1.Evidence" has at least the
	// Equivocation implementations.
	resImpl, err := s.queryClient.ListImplementations(
		context.Background(),
		&reflection.ListImplementationsRequest{InterfaceName: iface},
	)
	s.Require().NoError(err)
	s.Require().Contains(resImpl.GetImplementationMessageNames(), "/cosmos.evidence.v1beta1.Equivocation")
}

func TestSimulateTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
