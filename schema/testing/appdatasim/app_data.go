package appdatasim

import (
	"fmt"
	"sort"

	"pgregory.net/rapid"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
	schematesting "cosmossdk.io/schema/testing"
	"cosmossdk.io/schema/testing/statesim"
	"cosmossdk.io/schema/view"
)

// Options are the options for creating an app data simulator.
type Options struct {
	// AppSchema is the schema to use. If it is nil, then schematesting.ExampleAppSchema
	// will be used.
	AppSchema map[string]schema.ModuleSchema

	// Listener is the listener to output appdata updates to.
	Listener appdata.Listener

	// StateSimOptions are the options to pass to the statesim.App instance used under
	// the hood.
	StateSimOptions statesim.Options
}

// Simulator simulates a stream of app data. Currently, it only simulates InitializeModuleData, OnObjectUpdate,
// StartBlock and Commit callbacks but others will be added in the future.
type Simulator struct {
	state    *statesim.App
	options  Options
	blockNum uint64
}

// BlockData represents the app data packets in a block.
type BlockData = []appdata.Packet

// NewSimulator creates a new app data simulator with the given options and runs its
// initialization methods.
func NewSimulator(options Options) (*Simulator, error) {
	sim := &Simulator{
		// we initialize the state simulator with no app schema because we'll do
		// that in the first block
		state:   statesim.NewApp(nil, options.StateSimOptions),
		options: options,
	}

	err := sim.initialize()
	if err != nil {
		return nil, err
	}

	return sim, nil
}

func (a *Simulator) initialize() error {
	// in block "0" we only pass module initialization data and don't
	// even generate any real block data
	var keys []string
	for key := range a.options.AppSchema {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, moduleName := range keys {
		err := a.ProcessPacket(appdata.ModuleInitializationData{
			ModuleName: moduleName,
			Schema:     a.options.AppSchema[moduleName],
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// BlockDataGen generates random block data. It is expected that generated data is passed to ProcessBlockData
// to simulate the app data stream and advance app state based on the object updates in the block. The first
// packet in the block data will be a StartBlockData packet with the height set to the next block height.
func (a *Simulator) BlockDataGen() *rapid.Generator[BlockData] {
	return a.BlockDataGenN(0, 100)
}

// BlockDataGenN creates a block data generator which allows specifying the maximum number of updates per block.
func (a *Simulator) BlockDataGenN(minUpdatesPerBlock, maxUpdatesPerBlock int) *rapid.Generator[BlockData] {
	numUpdatesGen := rapid.IntRange(minUpdatesPerBlock, maxUpdatesPerBlock)

	return rapid.Custom(func(t *rapid.T) BlockData {
		var packets BlockData

		packets = append(packets, appdata.StartBlockData{Height: a.blockNum + 1})

		updateSet := map[string]bool{}
		// filter out any updates to the same key from this block, otherwise we can end up with hard to debug errors
		updateGen := a.state.UpdateGen().Filter(func(data appdata.ObjectUpdateData) bool {
			for _, update := range data.Updates {
				_, existing := updateSet[a.formatUpdateKey(data.ModuleName, update)]
				if existing {
					return false
				}
			}
			return true
		})
		numUpdates := numUpdatesGen.Draw(t, "numUpdates")
		for i := 0; i < numUpdates; i++ {
			data := updateGen.Draw(t, fmt.Sprintf("update[%d]", i))
			for _, update := range data.Updates {
				// we need to set the update here each time so that this is used to filter out duplicates in the next round
				updateSet[a.formatUpdateKey(data.ModuleName, update)] = true
			}
			packets = append(packets, data)
		}

		packets = append(packets, appdata.CommitData{})

		return packets
	})
}

func (a *Simulator) formatUpdateKey(moduleName string, update schema.StateObjectUpdate) string {
	mod, err := a.state.GetModule(moduleName)
	if err != nil {
		panic(err)
	}

	objColl, err := mod.GetObjectCollection(update.TypeName)
	if err != nil {
		panic(err)
	}

	ks := fmt.Sprintf("%s:%s:%s", moduleName, update.TypeName, schematesting.ObjectKeyString(objColl.ObjectType(), update.Key))
	return ks
}

// ProcessBlockData processes the given block data, advancing the app state based on the object updates in the block
// and forwarding all packets to the attached listener. It is expected that the data passed came from BlockDataGen,
// however, other data can be passed as long as any StartBlockData packet has the height set to the current height + 1.
func (a *Simulator) ProcessBlockData(data BlockData) error {
	for _, packet := range data {
		err := a.ProcessPacket(packet)
		if err != nil {
			return err
		}
	}
	return nil
}

// ProcessPacket processes a single packet, advancing the app state based on the data in the packet,
// and forwarding the packet to any listener.
func (a *Simulator) ProcessPacket(packet appdata.Packet) error {
	err := a.options.Listener.SendPacket(packet)
	if err != nil {
		return err
	}
	switch packet := packet.(type) {
	case appdata.StartBlockData:
		if packet.Height != a.blockNum+1 {
			return fmt.Errorf("invalid StartBlockData packet: %v", packet)
		}
		a.blockNum++
	case appdata.ModuleInitializationData:
		return a.state.InitializeModule(packet)
	case appdata.ObjectUpdateData:
		return a.state.ApplyUpdate(packet)
	}
	return nil
}

// AppState returns the current app state backing the simulator.
func (a *Simulator) AppState() view.AppState {
	return a.state
}

// BlockNum returns the current block number of the simulator.
func (a *Simulator) BlockNum() (uint64, error) {
	return a.blockNum, nil
}
