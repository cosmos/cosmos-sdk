package baseapp

import abci "github.com/cometbft/cometbft/abci/types"

type StreamEvents struct {
	Events []abci.Event
	Height uint64
	Flush  bool
}

func (app *BaseApp) AddStreamEvents(height int64, events []abci.Event, flush bool) {
	go func() {
		if app.EnableStreamer {
			app.StreamEvents <- StreamEvents{
				Events: events,
				Height: uint64(height),
				Flush:  flush,
			}
		}
	}()
}
