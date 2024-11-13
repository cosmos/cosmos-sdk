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
	errExit          bool
}

func NewKeeper(kvMap KVServiceMap, telemetryService telemetry.Service) *Keeper {
	return &Keeper{
		kvServiceMap:     kvMap,
		telemetryService: telemetryService,
		validate:         false,
		errExit:          false,
	}
}

func (k *Keeper) LoadTest(ctx context.Context, msg *benchmark.MsgLoadTest) (*benchmark.MsgLoadTestResponse, error) {
	res := &benchmark.MsgLoadTestResponse{}
	for _, op := range msg.Ops {
		k.telemetryService.IncrCounter([]string{"benchmark", "op", "cnt"}, 1)
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

func (k *Keeper) countMiss(opType string) {
	k.telemetryService.IncrCounter([]string{"benchmark", "miss"}, 1, telemetry.Label{Name: "op", Value: opType})
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
		if k.validate {
			exists, err := kv.Has(key)
			if err != nil {
				return err
			}
			if !exists {
				k.countMiss("delete")
				if k.errExit {
					return fmt.Errorf("key %d not found", op.Seed)
				}
			}
		}
		return kv.Delete(key)
	case op.ValueLength > 0:
		opType := "insert"
		if op.Exists {
			opType = "update"
		}
		defer k.measureSince(start, opType)
		if k.validate {
			exists, err := kv.Has(key)
			if err != nil {
				return err
			}
			if exists != op.Exists {
				k.countMiss(opType)
				if k.errExit {
					return fmt.Errorf("key %d exists=%t, expected=%t", op.Seed, exists, op.Exists)
				}
			}
		}
		value := gen.Bytes(op.Seed, op.ValueLength)
		return kv.Set(key, value)
	case op.Iterations > 0:
		return fmt.Errorf("iterator not implemented")
	case op.ValueLength == 0:
		defer k.measureSince(start, "get")
		v, err := kv.Get(key)
		if v == nil {
			// always count a miss on GET since it requires no extra I/O
			k.countMiss("get")
			if k.errExit {
				return fmt.Errorf("key %s not found", key)
			}
		}
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
