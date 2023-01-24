package keeper

import (
	"testing"

	"github.com/regen-network/gocuke"
	"gotest.tools/v3/assert"
)

func TestAuthorize(t *testing.T) {
	t.Skip("TODO: uncomment this after implementing")
	gocuke.NewRunner(t, &authorizeSuite{}).Path("../features/msg_authorize.feature").Run()
}

type authorizeSuite struct {
	t   gocuke.TestingT
	err error
}

func (s *authorizeSuite) HasPermission(a string, b string) {
	panic("PENDING")
}

func (s *authorizeSuite) HasNoPermissions(a string) {
	panic("PENDING")
}

func (s *authorizeSuite) AttemptsToGrantThePermissions(a string, b string, c gocuke.DocString) {
	panic("PENDING")
}

func (s *authorizeSuite) ExpectSuccess() {
	assert.NoError(s.t, s.err)
}

func (s *authorizeSuite) ExpectAnError(a string) {
	assert.Error(s.t, s.err)
}

func (s *authorizeSuite) ExpectThatHasNoPermissions(a string) {
	panic("PENDING")
}
