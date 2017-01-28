# Develop and run a basecoin plugin

Basecoin is a demonstration environment of a Tendermint powered blockchain crypto-currency. This tutorial will walk you through of how to write a basic plugin for basecoin and deploy your own instance of basecoin running your plugin.  

A completed copy of the source code described in this tutorial can be found [here][1]

[1]: https://github.com/tendermint/basecoin-examples/blob/docs/pluginDev/

### Prerequisite installation

1. Make sure you [have Go installed][2] and [put $GOPATH/bin in your $PATH][3]
2. Download basecoin using `go get github.com/tendermint/basecoin`
3. Download basecoin's dependencies using `make get_vendor_deps`
3. Install basecoin using `make install`

[2]: https://golang.org/doc/install
[3]: https://github.com/tendermint/tendermint/wiki/Setting-GOPATH 
 
### Programming a custom basecoin application

The first step is to create a new project under your github directory, start by naming your project bcTutorial  
`mkdir $GOPATH/src/username/bcTutorial`  
Here _username_ should be replaced with your github username.  

Create a new file under your new directory with the file name main.go. Within main.go add the following information

	package main
	
	import (

	)
	
	func main() {
	}

Under the main function initialize a new instance of basecoin, this is done by first connecting to a new [merkleeyes][4] client and further initializing basecoin with the merkleeyes client. Add the following lines to your import block: 
[4]: https://github.com/tendermint/merkleeyes

	import(
		"flag"
		"github.com/tendermint/basecoin/app"
		eyes "github.com/tendermint/merkleeyes/client"
		cmn "github.com/tendermint/go-common"
	)

Add the following lines to your main function
	
	//flags
	eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or 'local' for embedded")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr, "socket")
	if err != nil {
		cmn.Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin
	bcApp := app.NewBasecoin(eyesCli)

Next, set the initial state of the basecoin app load the genesis file (if any) for basecoin. Your main function should now look as follows

	//flags
	eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or 'local' for embedded")
	genFilePath := flag.String("genesis", "", "Genesis file, if any")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr, "socket")
	if err != nil {
		cmn.Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin
	bcApp := app.NewBasecoin(eyesCli)

	// If genesis file was specified, set key-value options
	if *genFilePath != "" {
		err := bcApp.LoadGenesis(*genFilePath)
		if err != nil {
			cmn.Exit(cmn.Fmt("%+v", err))
		}
	}

Finally implement an abci listening server for the basecoin app. Within your imports block add the following line
	
	"github.com/tendermint/abci/server"

Within the main function add the listener, your main function should now look as follows
 
	//flags
	addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen address")
	eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or 'local' for embedded")
	genFilePath := flag.String("genesis", "", "Genesis file, if any")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr, "socket")
	if err != nil {
		cmn.Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin
	bcApp := app.NewBasecoin(eyesCli)

	// If genesis file was specified, set key-value options
	if *genFilePath != "" {
		err := bcApp.LoadGenesis(*genFilePath)
		if err != nil {
			cmn.Exit(cmn.Fmt("%+v", err))
		}
	}

	// Start the listener
	svr, err := server.NewServer(*addrPtr, "socket", bcApp)
	if err != nil {
		cmn.Exit("create listener: " + err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})

Great! Your application should be able to install and run, although you haven't yet added a plugin yet.

### Programming a basic plugin
 
The next step is to actually make your plugin that can be run your custom instance of basecoin. First create a directory and an empty .go with the path `plugins/counter/counter.go`. By definition any basecoin plugin needs to satisfy the basecoin [Plugin interface][5]. The following is a description of the required functions:
[5]: https://github.com/tendermint/basecoin/blob/master/types/plugin.go#L8-L15

 - Name() string
   - return the name of the application
 - SetOption(store KVStore, key string, value string) (log string)
   - SetOption may be called during genesis of basecoin and can be used to set initial plugin parameters 
 - RunTx(store KVStore, ctx CallContext, txBytes []byte) (res abci.Result)
   - RunTx contains the core logic of the plugin. This function is called during operation a plugins transaction is called 
 - InitChain(store KVStore, vals []\*abci.Validator)
   - Not-used within this tutorial
 - BeginBlock(store KVStore, height uint64)
   - Not-used within this tutorial
 - EndBlock(store KVStore, height uint64) []\*abci.Validator
   - Not-used within this tutorial

For this example SetOption, InitChain, BeginBlock, and EndBlock are not used, but are still present to satisfy the basecoin Plugin interface. To begin add the following basic information to counter.go

	package counter
	
	import (
		abci "github.com/tendermint/abci/types"
	)

	type CounterPlugin struct {
	}
	
	func (cp *CounterPlugin) Name() string {
		return ""
	}

	func (cp *CounterPlugin) SetOption(store types.KVStore, key string, value string) (log string) {
		return ""
	}
	
	func (cp *CounterPlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
		return abci.OK
	}
	
	func (cp *CounterPlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
	}
	
	func (cp *CounterPlugin) BeginBlock(store types.KVStore, height uint64) {
	}
	
	func (cp *CounterPlugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
		return nil
	}

Next, modify the Name() function to return a new variable that will be held within the CounterPlugin struct:

	type CounterPlugin struct {
		name string
	}
	
	func (cp *CounterPlugin) Name() string {
		return cp.name
	}

Add a new exposed function to initialize a new CounterPlugin instance

	func New(name string) *CounterPlugin {
		return &CounterPlugin{
			name: name,
		}
	}

We now can move onto programming the core tx logic within the RunTx function. RunTx contains three input fields:
 - store types.KVStore
   - This term provides read/write capabilities to the merkelized data store which is accessible cross-plugin
 - ctx types.CallContext
   - The ctx contains the callers address, a pointer to the callers account, and an amount of coins sent with the transaction
 - txBytes []byte
   - This field can be used to send customized information from the basecoin application to your plugin

To begin we will define and implement customized information to be read in from txBytes. This is accomplished by creating a new struct containing all the desired terms to be passed through the txBytes term and further decode the information from this term with [go-wire][6]. Note that because we are decoding txBytes, all transactions to our plugin will also need to contain the txBytes term as encoded by go-wire.
[6]: https://github.com/tendermint/go-wire

First add a new struct to counter.go. We will include two custom terms for this exercise:
 - Valid: should the transaction be performed 
 - Cost: cost that the CounterTx Plugin charges per interaction  

	type CounterTx struct {
		Valid bool
		Cost  types.Coins
	}

Next, we need to import the go-wire library. Update the import section to the following

	import (
		abci "github.com/tendermint/abci/types"
		"github.com/tendermint/go-wire"
	)

Next within the RunTx function add the following code to decode txBytes into a new instance of CounterTx

	// Decode tx
	var tx CounterTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

With all the basic information loaded in RunTx, we can perform some basic checks. First check the _Valid_ term and terminate the process if it's set to false. 

	// Validate tx
	if !tx.Valid {
		return abci.ErrInternalError.AppendLog("CounterTx.Valid must be true")
	}

Further check if the cost term is valid and non-negative.

	if !tx.Cost.IsValid() {
		return abci.ErrInternalError.AppendLog("CounterTx.Cost is not sorted or has zero amounts")
	}
	if !tx.Cost.IsNonnegative() {
		return abci.ErrInternalError.AppendLog("CounterTx.Cost must be nonnegative")
	}

Check if the cost is been covered by the amount of coins sent through the context (_ctx.Coins_) 

	// Did the caller provide enough coins?
	if !ctx.Coins.IsGTE(tx.Cost) {
		return abci.ErrInsufficientFunds.AppendLog("CounterTx.Cost was not provided")
	}

Now that the checks are complete, the core-calculations can be calculated and saved with basecoin. The results of the core-calculation can be saved as to a new 'state' struct, encoded to a byte array using go-wire, and saved to basecoin using the KVStore term passed to RunTx through _store_ term. For this tutorial we will choose our core-calculation terms to be a count for the number of transactions, as well as the total amount paid to the counter plugin through transactions. To begin, define a new state struct  

	type CounterPluginState struct {
		Counter   int
		TotalCost types.Coins
	}

Next, within the RunTx function load the state if exists. Note that if the state does not exist, the state bytes read into _cpStateBytes_ will have no members therefore returning a length of zero. If this is the case we will skip loading the initial state  

	// Load CounterPluginState
	var cpState CounterPluginState
	cpStateBytes := store.Get(cp.StateKey())
	if len(cpStateBytes) > 0 {
		err = wire.ReadBinaryBytes(cpStateBytes, &cpState)
		if err != nil {
			return abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
		}
	}

Now we can update the values of the state variables to reflect the ongoing transaction, and save the transaction results back to the store. 
	
	// Update CounterPluginState
	cpState.Counter += 1
	cpState.TotalCost = cpState.TotalCost.Plus(tx.Cost)

	// Save CounterPluginState
	store.Set(cp.StateKey(), wire.BinaryBytes(cpState))

Our final RunTx method should now look like this:

	func (cp *CounterPlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
		// Decode tx
		var tx CounterTx
		err := wire.ReadBinaryBytes(txBytes, &tx)
		if err != nil {
			return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
		}
	
		// Validate tx
		if !tx.Valid {
			return abci.ErrInternalError.AppendLog("CounterTx.Valid must be true")
		}
		if !tx.Cost.IsValid() {
			return abci.ErrInternalError.AppendLog("CounterTx.Cost is not sorted or has zero amounts")
		}
		if !tx.Cost.IsNonnegative() {
			return abci.ErrInternalError.AppendLog("CounterTx.Cost must be nonnegative")
		}
	
		// Did the caller provide enough coins?
		if !ctx.Coins.IsGTE(tx.Cost) {
			return abci.ErrInsufficientFunds.AppendLog("CounterTx.Cost was not provided")
		}
	
		// Load CounterPluginState
		var cpState CounterPluginState
		cpStateBytes := store.Get(cp.StateKey())
		if len(cpStateBytes) > 0 {
			err = wire.ReadBinaryBytes(cpStateBytes, &cpState)
			if err != nil {
				return abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
			}
		}
	
		// Update CounterPluginState
		cpState.Counter += 1
		cpState.TotalCost = cpState.TotalCost.Plus(tx.Cost)
	
		// Save CounterPluginState
		store.Set(cp.StateKey(), wire.BinaryBytes(cpState))
	
		return abci.OK
	}


Lastly, we can implement our plugin into our main.go method. Within the main function define and register your plugin with the basecoin instance, after you have defined the instance but before you have run the Tendermint listening server. 

	// create/add plugins
	counter := counter.New("counter")
	bcApp.RegisterPlugin(counter)

In order for this code to work add the following line to your import block, where _username_ is replaced with your username.
	
	"github.com/username/bcTutorial/plugins/counter"

Our main function within main.go should now look as follows

	func main() {
	
		//flags
		addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen address")
		eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or 'local' for embedded")
		genFilePath := flag.String("genesis", "", "Genesis file, if any")
		flag.Parse()
	
		// Connect to MerkleEyes
		eyesCli, err := eyes.NewClient(*eyesPtr, "socket")
		if err != nil {
			cmn.Exit("connect to MerkleEyes: " + err.Error())
		}
	
		// Create Basecoin
		bcApp := app.NewBasecoin(eyesCli)
	
		// create/add plugins
		counter := counter.New("counter")
		bcApp.RegisterPlugin(counter)
	
		// If genesis file was specified, set key-value options
		if *genFilePath != "" {
			err := bcApp.LoadGenesis(*genFilePath)
			if err != nil {
				cmn.Exit(cmn.Fmt("%+v", err))
			}
		}
	
		// Start the listener
		svr, err := server.NewServer(*addrPtr, "socket", bcApp)
		if err != nil {
			cmn.Exit("create listener: " + err.Error())
		}
	
		// Wait forever
		cmn.TrapSignal(func() {
			// Cleanup
			svr.Stop()
		})
	}

A completed copy of the source code described in this tutorial can be found [here][1]

### Installing and running your application

Your application can be installed using the command `go install` when navigated to the root directory of your application folder. 
To run your new application use the following steps:

1. In a terminal window run `counter`
2. Within a second terminal window run `tendermint node`

Your blockchain basecoin application is now operating and can be communicated with through RPC calls!
