package keeper

import (
	"testing"

	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
)

func TestReset(t *testing.T) {
	t.Skip("TODO: uncomment this after implementing")
	gocuke.NewRunner(t, &resetSuite{}).Path("../features/msg_authorize.feature").Run()
}

type resetSuite struct {
	t   gocuke.TestingT
	err error
}

func (s *resetSuite) HasPermission(a string, b string) {
	panic("PENDING")
}

func (s *resetSuite) HasNoPermissions(a string) {
	panic("PENDING")
}

func (s *resetSuite) AttemptsToResetCircuit(a string, b string, c gocuke.DocString) {
	panic("PENDING")
}

func (s *resetSuite) ExpectSuccess() {
	require.NoError(s.t, s.err)
}

func (s *resetSuite) ExpectAnError(a string) {
	require.EqualError(s.t, s.err, a)
}

func (s *resetSuite) ExpectThatHasNoPermissions(a string) {
	panic("PENDING")
}
