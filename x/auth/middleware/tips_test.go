package middleware_test

import (
	"testing"
)

func (s *MWTestSuite) TestMetaTransactions() {
	s.SetupTest(false) // reset

	tipper := s.createTestAccounts(1, "regen")[0]
}

func TestMWTestSuite2(t *testing.T) {
	s.Run(t, new(MWTestSuite))
}
