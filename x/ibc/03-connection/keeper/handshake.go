package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// // CONTRACT: client and remote must be filled by the caller
// func (k Keeper) CreateState(parent State) HandshakeState {
// 	return HandshakeState{
// 		State:              parent,
// 		Stage:              k.man.protocol.Value([]byte(parent.id + "/state")).Enum(),
// 		CounterpartyClient: k.man.protocol.Value([]byte(parent.id + "/counterpartyClient")).String(),
// 	}
// }

// func (man CounterpartyHandshaker) CreateState(id string) CounterHandshakeState {
// 	return CounterHandshakeState{
// 		CounterState:       k.man.CreateState(id),
// 		Stage:              k.man.protocol.Value([]byte(id + "/state")).Enum(),
// 		CounterpartyClient: k.man.protocol.Value([]byte(id + "/counterpartyClient")).String(),
// 	}
// }

// func (k Keeper) create(ctx sdk.Context, id string, connection Connection, counterpartyClient string) (obj HandshakeState, err error) {
// 	cobj, err := k.man.create(ctx, id, connection, HandshakeKind)
// 	if err != nil {
// 		return
// 	}
// 	obj = k.CreateState(cobj)
// 	obj.CounterpartyClient.Set(ctx, counterpartyClient)
// 	obj.Counterparty = k.CounterParty.CreateState(connection.Counterparty)
// 	return obj, nil
// }

// func (k Keeper) query(ctx sdk.Context, id string) (obj HandshakeState, err error) {
// 	cobj, err := k.man.query(ctx, id, HandshakeKind)
// 	if err != nil {
// 		return
// 	}
// 	obj = k.CreateState(cobj)
// 	obj.Counterparty = k.counterParty.CreateState(obj.GetConnection(ctx).Counterparty)
// 	return
// }

// ConnOpenInit initialises a connection attempt on chain A.
func (k Keeper) ConnOpenInit(
	ctx sdk.Context, connectionID, clientID string, counterparty types.Counterparty,
) error {
	_, found := k.GetConnection(ctx, connectionID)
	if found {
		return sdkerrors.Wrap(types.ErrConnectionExists(k.codespace), "cannot initialize connection")
	}

	connection := types.NewConnectionEnd(clientID, counterparty, k.getCompatibleVersions())
	connection.State = types.INIT

	k.SetConnection(ctx, connectionID, connection)
	err := k.addConnectionToClient(ctx, clientID, connectionID)
	if err != nil {
		sdkerrors.Wrap(err, "cannot initialize connection")
	}

	return nil
}

// ConnOpenTry relays notice of a connection attempt on chain A to chain B (this
// code is executed on chain B).
func (k Keeper) ConnOpenTry(ctx sdk.Context, connectionID, clientID string, counterparty types.Counterparty,
	counterpartyVersions []string, proofInit ics23.Proof, proofHeight uint64, consensusHeight uint64,
) error {

	if consensusHeight > uint64(ctx.BlockHeight()) {
		return errors.New("invalid consensus height")
	}

	// obj, err = k.create(ctx, id, connection, counterpartyClient)
	// if err != nil {
	// 	return
	// }

	// ctx, err = k.Context(ctx, height, proofs)
	// if err != nil {
	// 	return
	// }

	// if !obj.Counterparty.Stage.Is(ctx, Init) {
	// 	err = errors.New("counterParty state not init")
	// 	return
	// }

	// if !obj.Counterparty.Connection.Is(ctx, Connection{
	// 	Client:       counterpartyClient,
	// 	Counterparty: id,
	// 	Path:         obj.path,
	// }) {
	// 	err = errors.New("wrong counterParty connection")
	// 	return
	// }

	// if !obj.Counterparty.CounterpartyClient.Is(ctx, connection.Client) {
	// 	err = errors.New("counterParty client not match")
	// 	return
	// }

	// // TODO: commented out, need to check whether the stored client is compatible
	// // make a separate module that manages recent n block headers
	// // ref #4647
	// /*
	// 	var expected client.ConsensusState
	// 	obj.self.Get(ctx, expheight, &expected)
	// 	if !obj.counterParty.client.Is(ctx, expected) {
	// 		return errors.New("unexpected counterParty client value")
	// 	}
	// */

	// // CONTRACT: OpenTry() should be called after man.Create(), not man.Query(),
	// // which will ensure
	// // assert(get("connections/{desiredIdentifier}") === null) and
	// // set("connections{identifier}", connection)

	// obj.Stage.Set(ctx, OpenTry)

	return nil
}

// ConnOpenAck relays acceptance of a connection open attempt from chain B back
// to chain A (this code is executed on chain A).
func (k Keeper) ConnOpenAck(ctx sdk.Context,
	proofs []ics23.Proof, height uint64,
	id string,
) error {
	// obj, err = man.query(ctx, id)
	// if err != nil {
	// 	return
	// }

	// ctx, err = obj.Context(ctx, height, proofs)
	// if err != nil {
	// 	return
	// }

	// if !obj.Stage.Transit(ctx, Init, Open) {
	// 	err = errors.New("ack on non-init connection")
	// 	return
	// }

	// if !obj.Counterparty.Connection.Is(ctx, Connection{
	// 	Client:       obj.CounterpartyClient.Get(ctx),
	// 	Counterparty: obj.ID(),
	// 	Path:         obj.path,
	// }) {
	// 	err = errors.New("wrong counterParty")
	// 	return
	// }

	// if !obj.Counterparty.Stage.Is(ctx, OpenTry) {
	// 	err = errors.New("counterParty state not opentry")
	// 	return
	// }

	// if !obj.Counterparty.CounterpartyClient.Is(ctx, obj.GetConnection(ctx).Client) {
	// 	err = errors.New("counterParty client not match")
	// 	return
	// }

	// // TODO: implement in v1
	// /*
	// 	var expected client.ConsensusState
	// 	// obj.self.Get(ctx, expheight, &expected)
	// 	if !obj.counterParty.client.Is(ctx, expected) {
	// 		// return errors.New("unexpected counterParty client value")
	// 	}
	// */
	// obj.Available.Set(ctx, true)

	return nil
}

// ConnOpenConfirm confirms opening of a connection on chain A to chain B, after
// which the connection is open on both chains (this code is executed on chain B).
func (k Keeper) ConnOpenConfirm(ctx sdk.Context,
	proofs []ics23.Proof, height uint64,
	id string) error {

	// obj, err = man.query(ctx, id)
	// if err != nil {
	// 	return
	// }

	// ctx, err = obj.Context(ctx, height, proofs)
	// if err != nil {
	// 	return
	// }

	// if !obj.Stage.Transit(ctx, OpenTry, Open) {
	// 	err = errors.New("confirm on non-try connection")
	// 	return
	// }

	// if !obj.Counterparty.Stage.Is(ctx, Open) {
	// 	err = errors.New("counterParty state not open")
	// 	return
	// }

	// obj.Available.Set(ctx, true)

	return nil
}
