package client

import (
	"errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type IDGenerator func(sdk.Context /*Header,*/, mapping.Value) string

func IntegerIDGenerator(ctx sdk.Context, v mapping.Value) string {
	id := mapping.NewInteger(v, mapping.Dec).Incr(ctx)
	return strconv.FormatUint(id, 10)
}

type Manager struct {
	protocol mapping.Mapping

	idval mapping.Value
	idgen IDGenerator
}

func NewManager(protocol, free mapping.Base, idgen IDGenerator) Manager {
	return Manager{
		protocol: mapping.NewMapping(protocol, []byte("/client")),
		idval:    mapping.NewValue(free, []byte("/client/id")),
		idgen:    idgen,
	}
}

type CounterpartyManager struct {
	protocol commitment.Mapping
}

func NewCounterpartyManager(protocol commitment.Base) CounterpartyManager {
	return CounterpartyManager{
		protocol: commitment.NewMapping(protocol, []byte("/client")),
	}
}

/*
func (man Manager) RegisterKind(kind Kind, pred ValidityPredicate) Manager {
	if _, ok := man.pred[kind]; ok {
		panic("Kind already registered")
	}
	man.pred[kind] = pred
	return man
}
*/
func (man Manager) object(id string) Object {
	return Object{
		id:     id,
		client: man.protocol.Value([]byte(id)),
		freeze: mapping.NewBoolean(man.protocol.Value([]byte(id + "/freeze"))),
	}
}

func (man Manager) Create(ctx sdk.Context, cs Client) string {
	id := man.idgen(ctx, man.idval)
	err := man.object(id).create(ctx, cs)
	if err != nil {
		panic(err)
	}
	return id
}

func (man Manager) Query(ctx sdk.Context, id string) (Object, error) {
	res := man.object(id)
	if !res.exists(ctx) {
		return Object{}, errors.New("client not exists")
	}
	return res, nil
}

type Object struct {
	id     string
	client mapping.Value
	freeze mapping.Boolean
}

func (obj Object) create(ctx sdk.Context, st Client) error {
	if obj.exists(ctx) {
		return errors.New("Create client on already existing id")
	}
	obj.client.Set(ctx, st)
	return nil
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

func (obj Object) Is(ctx sdk.Context, client Client) bool {
	return obj.client.Is(ctx, client)
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
	updated, err := stored.Validate(header)
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
