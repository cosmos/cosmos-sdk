package client

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// Any actor holding the Manager can access on and modify any client information
type Manager struct {
	protocol state.Mapping
}

func NewManager(base state.Mapping) Manager {
	return Manager{
		protocol: base.Prefix(LocalRoot()),
	}
}

type CounterpartyManager struct {
	protocol commitment.Mapping
}

func NewCounterpartyManager(cdc *codec.Codec) CounterpartyManager {
	return CounterpartyManager{
		protocol: commitment.NewMapping(cdc, LocalRoot()),
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
func (man Manager) Object(id string) Object {
	return Object{
		id:             id,
		Roots:          man.protocol.Prefix([]byte(id + "/roots/")).Indexer(state.Dec),
		ConsensusState: man.protocol.Value([]byte(id)),
		Frozen:         man.protocol.Value([]byte(id + "/freeze")).Boolean(),
	}
}

func (man Manager) Create(ctx sdk.Context, id string, cs ConsensusState) (Object, error) {
	obj := man.Object(id)
	if obj.exists(ctx) {
		return Object{}, errors.New("Create client on already existing id")
	}
	obj.Roots.Set(ctx, cs.GetHeight(), cs.GetRoot())
	obj.ConsensusState.Set(ctx, cs)
	return obj, nil
}

func (man Manager) Query(ctx sdk.Context, id string) (Object, error) {
	res := man.Object(id)
	if !res.exists(ctx) {
		return Object{}, errors.New("client not exists")
	}
	return res, nil
}

func (man CounterpartyManager) Object(id string) CounterObject {
	return CounterObject{
		id:             id,
		ConsensusState: man.protocol.Value([]byte(id)),
	}
}

func (man CounterpartyManager) Query(id string) CounterObject {
	return man.Object(id)
}

// Any actor holding the Object can access on and modify that client information
type Object struct {
	id             string
	Roots          state.Indexer
	ConsensusState state.Value // ConsensusState
	Frozen         state.Boolean
}

type CounterObject struct {
	id             string
	ConsensusState commitment.Value
}

func (obj Object) ID() string {
	return obj.id
}

func (obj Object) GetConsensusState(ctx sdk.Context) (res ConsensusState) {
	obj.ConsensusState.Get(ctx, &res)
	return
}

func (obj Object) GetRoot(ctx sdk.Context, height uint64) (res commitment.Root, err error) {
	err = obj.Roots.GetSafe(ctx, height, &res)
	return
}

func (obj CounterObject) Is(ctx sdk.Context, client ConsensusState) bool {
	return obj.ConsensusState.Is(ctx, client)
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.ConsensusState.Exists(ctx)
}

func (obj Object) Update(ctx sdk.Context, header Header) error {
	if !obj.exists(ctx) {
		panic("should not update nonexisting client")
	}

	if obj.Frozen.Get(ctx) {
		return errors.New("client is Frozen")
	}

	stored := obj.GetConsensusState(ctx)
	updated, err := stored.Validate(header)
	if err != nil {
		return err
	}

	obj.ConsensusState.Set(ctx, updated)
	obj.Roots.Set(ctx, updated.GetHeight(), updated.GetRoot())

	return nil
}

func (obj Object) Freeze(ctx sdk.Context) error {
	if !obj.exists(ctx) {
		panic("should not freeze nonexisting client")
	}

	if obj.Frozen.Get(ctx) {
		return errors.New("client is already Frozen")
	}

	obj.Frozen.Set(ctx, true)

	return nil
}

func (obj Object) Delete(ctx sdk.Context) error {
	if !obj.exists(ctx) {
		panic("should not delete nonexisting client")
	}

	if !obj.Frozen.Get(ctx) {
		return errors.New("client is not Frozen")
	}

	obj.ConsensusState.Delete(ctx)
	obj.Frozen.Delete(ctx)

	return nil
}
