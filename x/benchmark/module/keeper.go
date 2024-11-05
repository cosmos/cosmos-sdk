package benchmark

import (
	"context"
	"fmt"

	"cosmossdk.io/x/benchmark"
)

var _ benchmark.MsgServer = Keeper{}

type Keeper struct {
	collector *KVServiceCollector
}

func NewKeeper(collector *KVServiceCollector) *Keeper {
	return &Keeper{collector: collector}
}

// LoadTest implements MsgServer.
func (k Keeper) LoadTest(ctx context.Context, msg *benchmark.MsgLoadTest) (*benchmark.MsgLoadTestResponse, error) {
	res := &benchmark.MsgLoadTestResponse{}
	for _, op := range msg.Ops {
		_, ok := k.collector.services[op.Actor]
		if !ok {
			return res, fmt.Errorf("actor %s not found", op.Actor)
		}
	}
	return res, nil
}
