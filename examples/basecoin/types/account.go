package types

import (
	wire "github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

var _ sdk.Account = (*AppAccount)(nil)

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

type AppAccountCodec struct {
	cdc *wire.Codec
}

func NewAppAccountCodecFromWireCodec(cdc *wire.Codec) AppAccountCodec {
	return AppAccountCodec{cdc}
}

func (_ AppAccountCodec) Prototype() interface{} {
	return &AppAccount{}
}

func (aac AppAccountCodec) Encode(o interface{}) (bz []byte, err error) {
	return aac.cdc.MarshalBinary(o)
}

func (aac AppAccountCodec) Decode(bz []byte) (o interface{}, err error) {
	o = aac.Prototype()
	err = aac.cdc.UnmarshalBinary(bz, o)
	return o, err
}
