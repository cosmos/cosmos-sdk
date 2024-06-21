package listenertest

import (
	"context"
	"fmt"

	"pgregory.net/rapid"

	"cosmossdk.io/schema"
	schematesting "cosmossdk.io/schema/testing"
	"cosmossdk.io/schema/testing/statesim"
)

type AppSimulatorOptions struct {
	AppSchema          map[string]schema.ModuleSchema
	Listener           blockdata.Listener
	EventAlignedWrites bool
}

type AppDataSimulator struct {
	state        *statesim.App
	options      AppSimulatorOptions
	blockNum     uint64
	blockDataGen *rapid.Generator[BlockData]
}

type BlockData = []blockdata.Packet

func NewAppSimulator(options AppSimulatorOptions) *AppDataSimulator {
	if options.AppSchema == nil {
		options.AppSchema = schematesting.ExampleAppSchema
	}

	sim := &AppDataSimulator{
		options: options,
	}

	return sim
}

func (a *AppDataSimulator) Initialize() error {
	if f := a.options.Listener.InitializeModuleData; f != nil {
		err := a.state.ScanModuleSchemas(func(moduleName string, moduleSchema schema.ModuleSchema) error {
			return f(moduleName, moduleSchema)
		})
		if err != nil {
			return err
		}
	}

	if f := a.options.Listener.Initialize; f != nil {
		_, err := f(context.Background(), blockdata.InitializationData{
			HasEventAlignedWrites: a.options.EventAlignedWrites,
		})
		if err != nil {
			return err
		}
	}

	if f := a.options.Listener.CompleteInitialization; f != nil {
		err := f()
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (a *AppDataSimulator) BlockDataGen() *rapid.Generator[BlockData] {
	return a.BlockDataGenN(100)
}

func (a *AppDataSimulator) BlockDataGenN(maxUpdatesPerBlock int) *rapid.Generator[BlockData] {
	numUpdatesGen := rapid.IntRange(1, maxUpdatesPerBlock)

	return rapid.Custom(func(t *rapid.T) BlockData {
		var packets BlockData

		numUpdates := numUpdatesGen.Draw(t, "numUpdates")
		for i := 0; i < numUpdates; i++ {
			update := a.state.UpdateGen().Draw(t, fmt.Sprintf("update[%d]", i))
			packets = append(packets, update)
		}

		return packets
	})
}

func (a *AppDataSimulator) ProcessBlockData(data BlockData) error {
	a.blockNum++

	if f := a.options.Listener.StartBlock; f != nil {
		err := f(a.blockNum)
		if err != nil {
			return err
		}
	}

	for _, packet := range data {
		err := a.options.Listener.SendPacket(packet)
		if err != nil {
			return err
		}
	}

	if f := a.options.Listener.Commit; f != nil {
		err := f()
		if err != nil {
			return err
		}
	}

	return nil
}
