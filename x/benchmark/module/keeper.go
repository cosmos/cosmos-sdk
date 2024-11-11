package module

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/core/telemetry"
	"cosmossdk.io/x/benchmark"
	gen "cosmossdk.io/x/benchmark/generator"
)

var _ benchmark.MsgServer = &Keeper{}

type Keeper struct {
	kvServiceMap     KVServiceMap
	telemetryService telemetry.Service
	validate         bool
}

func NewKeeper(kvMap KVServiceMap, telemetryService telemetry.Service) *Keeper {
	return &Keeper{kvServiceMap: kvMap, telemetryService: telemetryService}
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

func (k *Keeper) measureSince(since time.Time, opType string) {
	k.telemetryService.MeasureSince(since, []string{"benchmark", "op"}, telemetry.Label{Name: "op", Value: opType})
}

func (k *Keeper) executeOp(ctx context.Context, op *benchmark.Op) error {
	start := time.Now()
	svc, ok := k.kvServiceMap[op.Actor]
	if !ok {
		return fmt.Errorf("actor %s not found", op.Actor)
	}
	kv := svc.OpenKVStore(ctx)
	key := gen.Bytes(op.Seed, op.KeyLength)
	switch {
	case op.Delete:
		defer k.measureSince(start, "delete")
		return kv.Delete(key)
	case op.ValueLength > 0:
		if k.validate {
			exists, err := kv.Has(key)
			if err != nil {
				return err
			}
			if exists != op.Exists {
				return fmt.Errorf("key %s exists=%t, expected=%t", key, exists, op.Exists)
			}
		}
		if op.Exists {
			defer k.measureSince(start, "update")
		} else {
			defer k.measureSince(start, "insert")
		}
		value := gen.Bytes(op.Seed, op.ValueLength)
		return kv.Set(key, value)
	case op.Iterations > 0:
		return fmt.Errorf("iterator not implemented")
	case op.ValueLength == 0:
		defer k.measureSince(start, "get")
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
