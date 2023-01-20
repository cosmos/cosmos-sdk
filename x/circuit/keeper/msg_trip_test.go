package keeper

import (
	"testing"

	"github.com/regen-network/gocuke"
)

func TestTrip(t *testing.T) {
	t.Skip("TODO: uncomment this after implementing")
	gocuke.NewRunner(t, &authorizeSuite{}).Path("../features/msg_authorize.feature").Run()
}

type tripSuite struct {
	t gocuke.TestingT
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
	panic("PENDING")
}

func (s *tripSuite) ExpectAnError(a string) {
	panic("PENDING")
}

func (s *tripSuite) ExpectThatHasNoPermissions(a string) {
	panic("PENDING")
}
