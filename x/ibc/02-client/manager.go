package client

import (
	"errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// XXX: implement spec: ClientState.verifiedRoots

type IDGenerator func(sdk.Context /*Header,*/, state.Value) string

func IntegerIDGenerator(ctx sdk.Context, v state.Value) string {
	id := state.NewInteger(v, state.Dec).Incr(ctx)
	return strconv.FormatUint(id, 10)
}

type Manager struct {
	protocol state.Mapping

	idval state.Value
	idgen IDGenerator
}

func NewManager(protocol, free state.Base, idgen IDGenerator) Manager {
	return Manager{
		protocol: state.NewMapping(protocol, []byte("/client")),
		idval:    state.NewValue(free, []byte("/client/id")),
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
		freeze: state.NewBoolean(man.protocol.Value([]byte(id + "/freeze"))),
	}
}

func (man Manager) Create(ctx sdk.Context, cs ConsensusState) (Object, error) {
	id := man.idgen(ctx, man.idval)
	obj := man.object(id)
	if obj.exists(ctx) {
		return Object{}, errors.New("Create client on already existing id")
	}
	obj.client.Set(ctx, cs)
	return obj, nil
}

func (man Manager) Query(ctx sdk.Context, id string) (Object, error) {
	res := man.object(id)
	if !res.exists(ctx) {
		return Object{}, errors.New("client not exists")
	}
	return res, nil
}

func (man CounterpartyManager) object(id string) CounterObject {
	return CounterObject{
		id:     id,
		client: man.protocol.Value([]byte(id)),
	}
}

func (man CounterpartyManager) Query(id string) CounterObject {
	return man.object(id)
}

type Object struct {
	id     string
	client state.Value // ConsensusState
	freeze state.Boolean
}

type CounterObject struct {
	id     string
	client commitment.Value
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.client.Exists(ctx)
}

func (obj Object) ID() string {
	return obj.id
}

func (obj Object) Value(ctx sdk.Context) (res ConsensusState) {
	obj.client.Get(ctx, &res)
	return
}

func (obj CounterObject) Is(ctx sdk.Context, client ConsensusState) bool {
	return obj.client.Is(ctx, client)
}

func (obj Object) Update(ctx sdk.Context, header Header) error {
	if !obj.exists(ctx) {
		panic("should not update nonexisting client")
	}

	if obj.freeze.Get(ctx) {
		return errors.New("client is frozen")
	}

	var stored ConsensusState
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
