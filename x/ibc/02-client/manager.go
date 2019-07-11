package client

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// XXX: implement spec: ClientState.verifiedRoots

type Manager struct {
	protocol state.Mapping

	idval state.Value
}

func NewManager(protocol, free state.Base) Manager {
	return Manager{
		protocol: state.NewMapping(protocol, []byte("/client")),
		idval:    state.NewValue(free, []byte("/client/id")),
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
		id:             id,
		consensusState: man.protocol.Value([]byte(id)),
		frozen:         state.NewBoolean(man.protocol.Value([]byte(id + "/freeze"))),
	}
}

func (man Manager) Create(ctx sdk.Context, id string, cs ConsensusState) (Object, error) {
	obj := man.object(id)
	if obj.exists(ctx) {
		return Object{}, errors.New("Create client on already existing id")
	}
	obj.consensusState.Set(ctx, cs)
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
		id:             id,
		consensusState: man.protocol.Value([]byte(id)),
	}
}

func (man CounterpartyManager) Query(id string) CounterObject {
	return man.object(id)
}

type Object struct {
	id             string
	consensusState state.Value // ConsensusState
	frozen         state.Boolean
}

type CounterObject struct {
	id             string
	consensusState commitment.Value
}

func (obj Object) ID() string {
	return obj.id
}

func (obj Object) ConsensusState(ctx sdk.Context) (res ConsensusState) {
	obj.consensusState.Get(ctx, &res)
	return
}

func (obj Object) Frozen(ctx sdk.Context) bool {
	return obj.frozen.Get(ctx)
}

func (obj CounterObject) Is(ctx sdk.Context, client ConsensusState) bool {
	return obj.consensusState.Is(ctx, client)
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.consensusState.Exists(ctx)
}

func (obj Object) Update(ctx sdk.Context, header Header) error {
	if !obj.exists(ctx) {
		panic("should not update nonexisting client")
	}

	if obj.Frozen(ctx) {
		return errors.New("client is frozen")
	}

	var stored ConsensusState
	obj.consensusState.Get(ctx, &stored)
	updated, err := stored.Validate(header)
	if err != nil {
		return err
	}

	obj.consensusState.Set(ctx, updated)

	return nil
}

func (obj Object) Freeze(ctx sdk.Context) error {
	if !obj.exists(ctx) {
		panic("should not freeze nonexisting client")
	}

	if obj.Frozen(ctx) {
		return errors.New("client is already frozen")
	}

	obj.frozen.Set(ctx, true)

	return nil
}

func (obj Object) Delete(ctx sdk.Context) error {
	if !obj.exists(ctx) {
		panic("should not delete nonexisting client")
	}

	if !obj.Frozen(ctx) {
		return errors.New("client is not frozen")
	}

	obj.consensusState.Delete(ctx)
	obj.frozen.Delete(ctx)

	return nil
}
