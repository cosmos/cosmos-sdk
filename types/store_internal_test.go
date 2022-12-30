package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type storeIntSuite struct {
	suite.Suite
}

func TestStoreIntSuite(t *testing.T) {
	suite.Run(t, new(storeIntSuite))
}
