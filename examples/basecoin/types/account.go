package types

import (
	"github.com/cosmos/cosmos-sdk/x/auth"
)

type AppAccount struct {
	auth.BaseAccount

	// Custom extensions for this application.
	Name string
}

func (acc AppAccount) GetName() string {
	return acc.Name
}

func (acc *AppAccount) SetName(name string) {
	acc.Name = name
}

//----------------------------------------

type AppAccountCodec struct{}

func (_ AppAccountCodec) Prototype() interface{} {
	return AppAccount{}
}

func (_ AppAccountCodec) Encode(o interface{}) (bz []byte, err error) {
	panic("not yet implemented")
}

func (_ AppAccountCodec) Decode(bz []byte) (o interface{}, err error) {
	panic("not yet implemented")
}
