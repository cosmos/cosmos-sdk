package keeper

import (
	"testing"

	"github.com/regen-network/gocuke"
	"gotest.tools/v3/assert"
)

func TestTrip(t *testing.T) {
	t.Skip("TODO: uncomment this after implementing")
	gocuke.NewRunner(t, &tripSuite{}).Path("../features/msg_trip.feature").Run()
}

type tripSuite struct {
	*baseFixture
}

func (s *tripSuite) Before(t *testing.T) {
	s.baseFixture = initFixture(t)
}

func (s *tripSuite) HasPermission(a, b string) {
	panic("PENDING")
}

func (s *tripSuite) HasNoPermissions(a string) {
	panic("PENDING")
}

func (s *tripSuite) AttemptsToTripCircuit(a, b string, c gocuke.DocString) {
	panic("PENDING")
}

func (s *tripSuite) ExpectSuccess() {
	assert.NilError(s.t, s.err)
}

func (s *tripSuite) ExpectAnError(a string) {
	assert.Error(s.t, s.err, a)
}

func (s *tripSuite) ExpectThatHasNoPermissions(a string) {
	panic("PENDING")
}
