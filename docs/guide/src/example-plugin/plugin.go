package main

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

//-----------------------------------------
// Structs
//   * Note the fields in each struct may be expanded/modified

// Plugin State Struct
//   * Intended to store the current state of the plugin
//     * This example contains a field which holds the execution count
//   * Serialized (by go-wire) and stored within the KVStore using the key retrieved
//     from the ExamplePlugin.StateKey() function/
//   * All fields must be exposed for serialization by external libs (here go-wire)
type ExamplePluginState struct {
	Counter int
}

// Transaction Struct
//   * Stores transaction-specific plugin-customized information
//     * This example contains a dummy field 'Valid' intended to specify
//       if the transaction is a valid and should proceed
//   * Deserialized (by go-wire) from txBytes in ExamplePlugin.RunTx
//   * All fields must be exposed for serialization by external libs (here go-wire)
type ExamplePluginTx struct {
	Valid bool
}

// Plugin Struct
//   * Struct which satisfies the basecoin Plugin interface
//   * Stores global plugin settings, in this example just the plugin name
type ExamplePlugin struct {
	name string
}

//-----------------------------------------
// Non-Mandatory Functions

// Return a new ExamplePlugin pointer with a hard-coded name
func NewExamplePlugin() *ExamplePlugin {
	return &ExamplePlugin{
		name: "example-plugin",
	}
}

// Return a byte array unique to this plugin which is used as the key
// to store the plugin state (ExamplePluginState)
func (ep *ExamplePlugin) StateKey() []byte {
	return []byte("ExamplePlugin.State")
}

//-----------------------------------------
// Basecoin Plugin Interface Functions

//Return the name of the plugin
func (ep *ExamplePlugin) Name() string {
	return ep.name
}

// SetOption may be called during genesis of basecoin and can be used to set
// initial plugin parameters. Within genesis.json file entries are made in
// the format: "<plugin>/<key>", "<value>" Where <plugin> is the plugin name,
// in this file ExamplePlugin.name, and <key> and <value> are the strings passed
// into the plugin SetOption function. This function is intended to be used to
// set plugin specific information such as the plugin state. Within this example
// SetOption is left unimplemented.
func (ep *ExamplePlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return ""
}

// The core tx logic of the app is containted within the RunTx function
// Input fields:
// - store types.KVStore
//   - This term provides read/write capabilities to the merkelized data store
//     which holds the basecoin state and is accessible to all plugins
// - ctx types.CallContext
//   - The ctx contains the callers address, a pointer to the callers account,
//     and an amount of coins sent with the transaction
// - txBytes []byte
//   - Used to send customized information to your plugin
//
// Other more complex plugins may have a variant on the process order within this
// example including loading and saving multiple or variable states, or not
// including a state stored in the KVStore whatsoever.
func (ep *ExamplePlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {

	// Decode txBytes using go-wire. Attempt to write the txBytes to the variable
	// tx, if the txBytes have not been properly encoded from a ExamplePluginTx
	// struct wire will produce an error.
	var tx ExamplePluginTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	// Perform Transaction Validation
	if !tx.Valid {
		return abci.ErrInternalError.AppendLog("Valid must be true")
	}

	// Load PluginState
	var pluginState ExamplePluginState
	stateBytes := store.Get(ep.StateKey())
	// If the state does not exist, stateBytes will be initialized
	// as an empty byte array with length of zero
	if len(stateBytes) > 0 {
		err = wire.ReadBinaryBytes(stateBytes, &pluginState) //decode using go-wire
		if err != nil {
			return abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
		}
	}

	//App Logic
	pluginState.Counter += 1

	// Save PluginState
	store.Set(ep.StateKey(), wire.BinaryBytes(pluginState))

	return abci.OK
}

func (ep *ExamplePlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
}

func (ep *ExamplePlugin) BeginBlock(store types.KVStore, hash []byte, header *abci.Header) {
}

func (ep *ExamplePlugin) EndBlock(store types.KVStore, height uint64) abci.ResponseEndBlock {
	return abci.ResponseEndBlock{}
}
