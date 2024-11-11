package module

import (
	"context"
	"fmt"

	"cosmossdk.io/x/benchmark"
	gen "cosmossdk.io/x/benchmark/generator"
)

var _ benchmark.MsgServer = &Keeper{}

type Keeper struct {
	kvServiceMap KVServiceMap
}

func NewKeeper(kvMap KVServiceMap) *Keeper {
	return &Keeper{kvServiceMap: kvMap}
}

func (k *Keeper) LoadTest(ctx context.Context, msg *benchmark.MsgLoadTest) (*benchmark.MsgLoadTestResponse, error) {
	res := &benchmark.MsgLoadTestResponse{}
	for _, op := range msg.Ops {
		err := k.executeOp(ctx, op)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func (k *Keeper) executeOp(ctx context.Context, op *benchmark.Op) error {
	svc, ok := k.kvServiceMap[op.Actor]
	if !ok {
		return fmt.Errorf("actor %s not found", op.Actor)
	}
	kv := svc.OpenKVStore(ctx)
	key := gen.Bytes(op.Seed, op.KeyLength)
	switch {
	case op.Delete:
		return kv.Delete(key)
	case op.ValueLength > 0:
		value := gen.Bytes(op.Seed, op.ValueLength)
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

func (k *Keeper) set(ctx context.Context, actor string, key, value []byte) error {
	svc, ok := k.kvServiceMap[actor]
	if !ok {
		return fmt.Errorf("actor %s not found", actor)
	}
	kv := svc.OpenKVStore(ctx)
	return kv.Set(key, value)
}
