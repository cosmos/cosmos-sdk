package appdatasim

import (
	"fmt"

	"pgregory.net/rapid"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
	schematesting "cosmossdk.io/schema/testing"
	"cosmossdk.io/schema/testing/statesim"
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

// Simulator simulates a stream of app data.
type Simulator struct {
	state        *statesim.App
	options      Options
	blockNum     uint64
	blockDataGen *rapid.Generator[BlockData]
}

// BlockData represents the app data packets in a block.
type BlockData = []appdata.Packet

// NewSimulator creates a new app data simulator with the given options.
func NewSimulator(options Options) *Simulator {
	if options.AppSchema == nil {
		options.AppSchema = schematesting.ExampleAppSchema
	}

	sim := &Simulator{
		state:   statesim.NewApp(options.AppSchema, options.StateSimOptions),
		options: options,
	}

	return sim
}

// Initialize runs the initialization methods of the app data stream.
func (a *Simulator) Initialize() error {
	if f := a.options.Listener.InitializeModuleData; f != nil {
		err := a.state.ScanModules(func(moduleName string, mod *statesim.Module) error {
			return f(appdata.ModuleInitializationData{ModuleName: moduleName, Schema: mod.ModuleSchema()})
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
	return a.BlockDataGenN(100)
}

// BlockDataGenN creates a block data generator which allows specifying the maximum number of updates per block.
func (a *Simulator) BlockDataGenN(maxUpdatesPerBlock int) *rapid.Generator[BlockData] {
	numUpdatesGen := rapid.IntRange(1, maxUpdatesPerBlock)

	return rapid.Custom(func(t *rapid.T) BlockData {
		var packets BlockData

		packets = append(packets, appdata.StartBlockData{Height: a.blockNum + 1})

		updateSet := map[string]bool{}
		// filter out any updates to the same key from this block, otherwise we can end up with weird errors
		updateGen := a.state.UpdateGen().Filter(func(data appdata.ObjectUpdateData) bool {
			for _, update := range data.Updates {
				_, existing := updateSet[fmt.Sprintf("%s:%v", data.ModuleName, update.Key)]
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
				updateSet[fmt.Sprintf("%s:%v", data.ModuleName, update.Key)] = true
			}
			packets = append(packets, data)
		}

		packets = append(packets, appdata.CommitData{})

		return packets
	})
}

// ProcessBlockData processes the given block data, advancing the app state based on the object updates in the block
// and forwarding all packets to the attached listener. It is expected that the data passed came from BlockDataGen,
// however, other data can be passed as long as the first packet is a StartBlockData packet with the block height
// set to the current block height + 1 and the last packet is a CommitData packet.
func (a *Simulator) ProcessBlockData(data BlockData) error {
	if len(data) < 2 {
		return fmt.Errorf("block data must contain at least two packets")
	}

	if startBlock, ok := data[0].(appdata.StartBlockData); !ok || startBlock.Height != a.blockNum+1 {
		return fmt.Errorf("first packet in block data must be a StartBlockData packet with height %d", a.blockNum+1)
	}

	if _, ok := data[len(data)-1].(appdata.CommitData); !ok {
		return fmt.Errorf("last packet in block data must be a CommitData packet")
	}

	// advance the block height
	a.blockNum++

	for _, packet := range data {
		// apply state updates from object updates
		if updateData, ok := packet.(appdata.ObjectUpdateData); ok {
			err := a.state.ApplyUpdate(updateData)
			if err != nil {
				return err
			}
		}

		// send the packet to the listener
		err := a.options.Listener.SendPacket(packet)
		if err != nil {
			return err
		}
	}

	return nil
}

// AppState returns the current app state backing the simulator.
func (a *Simulator) AppState() *statesim.App {
	return a.state
}

// BlockNum returns the current block number of the simulator.
func (a *Simulator) BlockNum() uint64 {
	return a.blockNum
}
