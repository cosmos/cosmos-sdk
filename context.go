package basecoin

import "github.com/tendermint/go-wire/data"

type Permission struct {
	App     string     // Which app authorized this?
	Address data.Bytes // App-specific identifier
}

func NewPermission(app string, addr []byte) Permission {
	return Permission{App: app, Address: addr}
}

// Context is an interface, so we can implement "secure" variants that
// rely on private fields to control the actions
type Context interface {
	// context.Context
	WithPermissions(perms ...Permission) Context
	HasPermission(perm Permission) bool
	IsParent(ctx Context) bool
	Reset() Context
}
