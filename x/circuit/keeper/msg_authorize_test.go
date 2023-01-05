package keeper

import (
	"testing"

	"github.com/regen-network/gocuke"
)

func TestAuthorize(t *testing.T) {
	t.Skip("TODO: uncomment this after implementing")
	gocuke.NewRunner(t, &authorizeSuite{}).Path("../features/msg_authorize.feature").Run()
}

type authorizeSuite struct {
	t gocuke.TestingT
}

func (s *authorizeSuite) ExpectSuccess() {
	panic("PENDING")
}

func (s *authorizeSuite) HasNoPermissions(a string) {
	panic("PENDING")
}

func (s *authorizeSuite) ExpectAnError(a string) {
	panic("PENDING")
}

func (s *authorizeSuite) ThatHas(a string, b string) {
	panic("PENDING")
}

func (s *authorizeSuite) ExpectSucesss() {
	panic("PENDING")
}

func (s *authorizeSuite) ExpectThatNoPermissions(a string) {
	panic("PENDING")
}

func (s *authorizeSuite) Has(a string, b string) {
	panic("PENDING")
}

func (s *authorizeSuite) AttemptsToGrantThePermissions(a string, b string, c gocuke.DocString) {
	panic("PENDING")
}
