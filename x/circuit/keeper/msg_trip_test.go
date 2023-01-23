package keeper

import (
	"testing"

	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"
)

func TestTrip(t *testing.T) {
	t.Skip("TODO: uncomment this after implementing")
	gocuke.NewRunner(t, &tripSuite{}).Path("../features/msg_authorize.feature").Run()
}

type tripSuite struct {
	t   gocuke.TestingT
	err error
}

func (s *tripSuite) HasPermission(a string, b string) {
	panic("PENDING")
}

func (s *tripSuite) HasNoPermissions(a string) {
	panic("PENDING")
}

func (s *tripSuite) AttemptsToTripCircuit(a string, b string, c gocuke.DocString) {
	panic("PENDING")
}

func (s *tripSuite) ExpectSuccess() {
	require.NoError(s.t, s.err)
}

func (s *tripSuite) ExpectAnError(a string) {
	require.EqualError(s.t, s.err, a)
}

func (s *tripSuite) ExpectThatHasNoPermissions(a string) {
	panic("PENDING")
}
