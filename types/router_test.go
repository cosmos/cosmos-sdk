package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type routeTestSuite struct {
	suite.Suite
}

func TestRouteTestSuite(t *testing.T) {
	suite.Run(t, new(routeTestSuite))
}

func (s *routeTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *routeTestSuite) TestNilRoute() {
	tests := []struct {
		name     string
		route    sdk.Route
		expected bool
	}{
		{
			name:     "all empty",
			route:    sdk.NewRoute("", nil),
			expected: true,
		},
		{
			name:     "only path",
			route:    sdk.NewRoute("some", nil),
			expected: true,
		},
		{
			name: "only handler",
			route: sdk.NewRoute("", func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
				return nil, nil
			}),
			expected: true,
		},
		{
			name: "correct route",
			route: sdk.NewRoute("some", func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
				return nil, nil
			}),
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Require().Equal(tt.expected, tt.route.Empty())
	}
}
