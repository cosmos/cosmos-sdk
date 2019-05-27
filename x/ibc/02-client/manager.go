package client

import (
	"errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IdentifierGenerator func(sdk.Context /*Header,*/, mapping.Value) string

func IntegerIdentifierGenerator(ctx sdk.Context, v mapping.Value) string {
	id := v.Integer().Incr(ctx)
	return strconv.FormatInt(id, 10)
}

type ClientManager struct {
	m     mapping.Mapping
	idval mapping.Value
	idgen IdentifierGenerator
	pred  map[ClientKind]ValidityPredicate
}

func NewClientManager(protocol, free mapping.Base, idgen IdentifierGenerator) *ClientManager {
	return &ClientManager{
		m:     mapping.NewMapping(protocol, nil),
		idval: mapping.NewValue(free, []byte{0x00}),
		idgen: idgen,
		pred:  make(map[ClientKind]ValidityPredicate),
	}
}

func (man ClientManager) RegisterKind(kind ClientKind, pred ValidityPredicate) ClientManager {
	if _, ok := man.pred[kind]; ok {
		panic("ClientKind already registered")
	}
	man.pred[kind] = pred
	return man
}

func (man ClientManager) object(key string) ClientObject {
	return ClientObject{
		key:    key,
		state:  man.m.Value([]byte("/" + key)),
		freeze: man.m.Value([]byte("/" + key + "/freeze")).Boolean(),
		pred:   man.pred,
	}
}

func (man ClientManager) Create(ctx sdk.Context, cs ConsensusState) string {
	key := man.idgen(ctx, man.idval)
	man.object(key).create(ctx, cs)
	return key
}

func (man ClientManager) Query(ctx sdk.Context, key string) (ClientObject, bool) {
	res := man.object(key)
	return res, res.exists(ctx)
}

type ClientObject struct {
	key    string
	state  mapping.Value
	freeze mapping.Boolean
	pred   map[ClientKind]ValidityPredicate
}

func (obj ClientObject) create(ctx sdk.Context, st ConsensusState) {
	if obj.exists(ctx) {
		panic("Create client on already existing key")
	}
	obj.state.Set(ctx, st)
}

func (obj ClientObject) exists(ctx sdk.Context) bool {
	return obj.state.Exists(ctx)
}

func (obj ClientObject) Update(ctx sdk.Context, header Header) error {
	if !obj.exists(ctx) {
		panic("should not update nonexisting client")
	}

	if obj.freeze.Get(ctx) {
		return errors.New("client is frozen")
	}

	var stored ConsensusState
	obj.state.GetIfExists(ctx, &stored)
	pred := obj.pred[stored.ClientKind()]

	updated, err := pred(stored, header)
	if err != nil {
		return err
	}

	obj.state.Set(ctx, updated)

	return nil
}

func (obj ClientObject) Freeze(ctx sdk.Context) error {
	if !obj.exists(ctx) {
		panic("should not freeze nonexisting client")
	}

	if obj.freeze.Get(ctx) {
		return errors.New("client is already frozen")
	}

	obj.freeze.Set(ctx, true)

	return nil
}

func (obj ClientObject) Delete(ctx sdk.Context) error {
	if !obj.exists(ctx) {
		panic("should not delete nonexisting client")
	}

	if !obj.freeze.Get(ctx) {
		return errors.New("client is not frozen")
	}

	obj.state.Delete(ctx)
	obj.freeze.Delete(ctx)

	return nil
}
