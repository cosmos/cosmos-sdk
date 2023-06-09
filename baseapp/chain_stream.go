package baseapp

import "github.com/cometbft/cometbft/abci/types"

type StreamEvents struct {
	Events []types.Event
	Height uint64
	Flush  bool
}

func (app *BaseApp) AddStreamEvents(height int64, events []types.Event, flush bool) {
	go func() {
		app.StreamEvents <- StreamEvents{
			Events: events,
			Height: uint64(height),
			Flush:  flush,
		}
	}()
}
