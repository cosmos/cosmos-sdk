package circuit

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestGRPCQueryTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCQueryTestSuite))
}
