package keeper

import (
	"testing"

	"github.com/regen-network/gocuke"
)

func TestReset(t *testing.T) {
	t.Skip("TODO: uncomment this after implementing")
	gocuke.NewRunner(t, &authorizeSuite{}).Path("../features/msg_authorize.feature").Run()
}

type resetSuite struct {
	t gocuke.TestingT
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
	panic("PENDING")
}

func (s *resetSuite) ExpectAnError(a string) {
	panic("PENDING")
}

func (s *resetSuite) ExpectThatHasNoPermissions(a string) {
	panic("PENDING")
}
