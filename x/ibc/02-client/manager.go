package client

import (
	"errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type IDGenerator func(sdk.Context /*Header,*/, mapping.Value) string

func IntegerIDGenerator(ctx sdk.Context, v mapping.Value) string {
	id := v.Integer().Incr(ctx)
	return strconv.FormatInt(id, 10)
}

type Manager struct {
	protocol mapping.Mapping

	idval mapping.Value
	idgen IDGenerator
	pred  map[Kind]ValidityPredicate
}

func NewManager(protocol, free mapping.Base, idgen IDGenerator) Manager {
	return Manager{
		protocol: mapping.NewMapping(protocol, []byte("/")),
		idval:    mapping.NewValue(free, []byte("/id")),
		idgen:    idgen,
		pred:     make(map[Kind]ValidityPredicate),
	}
}

func (man Manager) RegisterKind(kind Kind, pred ValidityPredicate) Manager {
	if _, ok := man.pred[kind]; ok {
		panic("ClientKind already registered")
	}
	man.pred[kind] = pred
	return man
}

func (man Manager) object(id string) Object {
	return Object{
		id:     id,
		client: man.protocol.Value([]byte(id)),
		freeze: man.protocol.Value([]byte(id + "/freeze")).Boolean(),
		pred:   man.pred,
	}
}

func (man Manager) Create(ctx sdk.Context, cs Client) string {
	id := man.idgen(ctx, man.idval)
	man.object(id).create(ctx, cs)
	return id
}

func (man Manager) Query(ctx sdk.Context, id string) (Object, bool) {
	res := man.object(id)
	return res, res.exists(ctx)
}

type Object struct {
	id     string
	client mapping.Value
	freeze mapping.Boolean
	pred   map[Kind]ValidityPredicate
}

func (obj Object) create(ctx sdk.Context, st Client) {
	if obj.exists(ctx) {
		panic("Create client on already existing id")
	}
	obj.client.Set(ctx, st)
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.client.Exists(ctx)
}

func (obj Object) ID() string {
	return obj.id
}

func (obj Object) Value(ctx sdk.Context) (res Client) {
	obj.client.Get(ctx, &res)
	return
}

func (obj Object) Update(ctx sdk.Context, header Header) error {
	if !obj.exists(ctx) {
		panic("should not update nonexisting client")
	}

	if obj.freeze.Get(ctx) {
		return errors.New("client is frozen")
	}

	var stored Client
	obj.client.GetIfExists(ctx, &stored)
	pred := obj.pred[stored.ClientKind()]

	updated, err := pred(stored, header)
	if err != nil {
		return err
	}

	obj.client.Set(ctx, updated)

	return nil
}

func (obj Object) Freeze(ctx sdk.Context) error {
	if !obj.exists(ctx) {
		panic("should not freeze nonexisting client")
	}

	if obj.freeze.Get(ctx) {
		return errors.New("client is already frozen")
	}

	obj.freeze.Set(ctx, true)

	return nil
}

func (obj Object) Delete(ctx sdk.Context) error {
	if !obj.exists(ctx) {
		panic("should not delete nonexisting client")
	}

	if !obj.freeze.Get(ctx) {
		return errors.New("client is not frozen")
	}

	obj.client.Delete(ctx)
	obj.freeze.Delete(ctx)

	return nil
}
