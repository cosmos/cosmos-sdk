package module

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/x/benchmark"
	"cosmossdk.io/x/benchmark/generator"
)

var _ benchmark.MsgServer = &Keeper{}

type Keeper struct {
	collector *KVServiceCollector
	generator *generator.Generator
}

func NewKeeper(collector *KVServiceCollector) *Keeper {
	return &Keeper{collector: collector}
}

func (k *Keeper) LoadTest(ctx context.Context, msg *benchmark.MsgLoadTest) (*benchmark.MsgLoadTestResponse, error) {
	res := &benchmark.MsgLoadTestResponse{}
	for _, op := range msg.Ops {
		svc, ok := k.collector.services[op.Actor]
		if !ok {
			return res, fmt.Errorf("actor %s not found", op.Actor)
		}
		kv := svc.OpenKVStore(ctx)
		err := k.executeOp(ctx, kv, op)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func (k *Keeper) executeOp(ctx context.Context, kv store.KVStore, op *benchmark.Op) error {
	key := k.generator.Bytes(op.Seed, op.KeyLength)
	switch {
	case op.Delete:
		return kv.Delete(key)
	case op.ValueLength > 0:
		value := k.generator.Bytes(op.Seed, op.ValueLength)
		return kv.Set(key, value)
	case op.Iterations > 0:
		return fmt.Errorf("iterator not implemented")
	case op.ValueLength == 0:
		_, err := kv.Get(key)
		return err
	default:
		return fmt.Errorf("invalid op: %+v", op)
	}
}
