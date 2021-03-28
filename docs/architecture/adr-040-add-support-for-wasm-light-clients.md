# ADR 39: Add support for Wasm based light client

## Changelog

- 26/11/2020: Initial Draft

## Status

Draft

## Abstract

In the Cosmos SDK light clients are current hardcoded in Go. This makes upgrading existing IBC light clients or add
support for new light client is multi step process which is time-consuming.

To remedy this, we are proposing a WASM VM to host light client bytecode, which allows easier upgrading of
existing IBC light clients as well as adding support for new IBC light clients without requiring a code release and corresponding
hard-fork event.

## Context
Currently in the SDK, light clients are defined as part of the codebase and are implemented as submodules under
`/x/ibc/light-clients/`.

Adding support for new light client or update an existing light client in the event of security
issue or consensus update is multi-step process:

1. Light client integration with SDK: IBC light clients are defined as part of the codebase and are implemented as
submodules under `/x/ibc/light-clients/`. To add support for new light client or update an existing light client in the
event of security issue or consensus update, we need to modify the codebase and integrate it in numerous places.

2. Governance voting: Adding new light client implementations require governance support and is expensive: This is
not ideal as chain governance is gatekeeper for new light client implementations getting added. If a small community
want support for light client X, they may not be able to convince governance to support it.

3. Validator upgrade: After governance voting succeeds, validators need to upgrade their nodes in order to enable new
IBC light client implementation. This is both time consuming and error prone.
   
Another problem stemming from the above process is that if a chain want to upgrade its own consensus, it will need to convince every chain
or hub connected to it to upgrade its light client in order to stay connected. Due to time consuming process required
to upgrade light client, a chain with lots of connections need to be disconnected for quite some time after upgrading 
its consensus, which can be very expensive.

We are proposing simplifying this workflow by integrating a WASM light client module which makes adding support for
a new light client a simple transaction. The light client bytecode, written in Wasm-compilable Rust, runs inside a Wasmer
VM. The Wasm light client submodule exposes a proxy light client interface that routes incoming messages to the
appropriate handler function, inside the Wasm VM for execution.

With WASM light client module, anybody can add new IBC light client in the form of WASM bytecode (provided they are able to pay the requisite gas fee for the transaction)
as well as instantiate clients using any created client type. This allows any chain to update its own light client in other chains
without going through steps outlined above.


## Decision

We decided to use WASM light client module as a light client proxy which will interface with the actual light client
uploaded as WASM bytecode. This will require changing client selection method to allow any client if the client type
has prefix of `wasm/`.

```go
// IsAllowedClient checks if the given client type is registered on the allowlist.
func (p Params) IsAllowedClient(clientType string) bool {
	for _, allowedClient := range p.AllowedClients {
		if allowedClient == clientType {
			return true
		}
	}
	return false || (p.AreWASMClientsAllowed && isWASMClient(clientType))
}
```

To upload new light client, user need to create a transaction with Wasm byte code which will be
processed by IBC Wasm module.

```go
func (k Keeper) UploadLightClient (wasmCode: []byte, description: String) {
    wasmRegistry = getWASMRegistry()
    id := hex.EncodeToString(sha256.Sum256(wasmCode))
    assert(!wasmRegistry.Exists(id))
    assert(wasmRegistry.ValidateAndStoreCode(id, description, wasmCode, false))
}
```

As name implies, Wasm registry is a registry which stores set of Wasm client code indexed by its hash and allows
client code to retrieve latest code uploaded.

`ValidateAndStoreCode` checks if the wasm bytecode uploaded is valid and confirms to VM interface.

### How light client proxy works?

The light client proxy behind the scenes will call a cosmwasm smart contract instance with incoming arguments in json
serialized format with appropriate environment information. Data returned by the smart contract is deserialized and
returned to the caller.

Consider an example of `CheckProposedHeaderAndUpdateState` function of `ClientState` interface. Incoming arguments are 
packaged inside a payload which is json serialized and passed to `callContract` which calls `vm.Execute` and returns the
array of bytes returned by the smart contract. This data is deserialized and passed as return argument.

```go
func (c *ClientState) CheckProposedHeaderAndUpdateState(context sdk.Context, marshaler codec.BinaryMarshaler, store sdk.KVStore, header exported.Header) (exported.ClientState, exported.ConsensusState, error) {
	// get consensus state corresponding to client state to check if the client is expired
	consensusState, err := GetConsensusState(store, marshaler, c.LatestHeight)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(
			err, "could not get consensus state from clientstore at height: %d", c.LatestHeight,
		)
	}
	
	payload := make(map[string]map[string]interface{})
	payload[CheckProposedHeaderAndUpdateState] = make(map[string]interface{})
	inner := payload[CheckProposedHeaderAndUpdateState]
	inner["me"] = c
	inner["header"] = header
	inner["consensus_state"] = consensusState

	encodedData, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(ErrUnableToMarshalPayload, fmt.Sprintf("underlying error: %s", err.Error()))
	}
	out, err := callContract(c.CodeId, context, store, encodedData)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(ErrUnableToCall, fmt.Sprintf("underlying error: %s", err.Error()))
	}
	output := clientStateCallResponse{}
	if err := json.Unmarshal(out.Data, &output); err != nil {
		return nil, nil, sdkerrors.Wrapf(ErrUnableToUnmarshalPayload, fmt.Sprintf("underlying error: %s", err.Error()))
	}
	if !output.Result.IsValid {
		return nil, nil, fmt.Errorf("%s error ocurred while updating client state", output.Result.ErrorMsg)
	}
	output.resetImmutables(c)
	return output.NewClientState, output.NewConsensusState, nil
}
```

## Consequences

### Positive
- Adding support for new light client or upgrading existing light client is way easier than before and only requires single transaction.
- Improves maintainability of Cosmos SDK, since no change in codebase is required to support new client or upgrade it.

### Negative
- Light clients need to be written in subset of rust which could compile in Wasm.
- Introspecting light client code is difficult as only compiled bytecode exists in the blockchain.
