package client

import (
	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ClientManager struct {
	m mapping.Mapping
}

func (man ClientManager) Create() ClientObject {

}

func (man ClientManager) Query(ctx sdk.Context, key []byte) ClientObject {

}

type ClientObject struct {
	key []byte
}

func (obj ClientObject) Update(ctx sdk.Context) {

}

func (obj ClientObject) Freeze(ctx sdk.Context) {

}

func (obj ClientObject) Delete() {

}
