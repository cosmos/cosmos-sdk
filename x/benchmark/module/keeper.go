package module

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/core/telemetry"
	"cosmossdk.io/x/benchmark"
	gen "cosmossdk.io/x/benchmark/generator"
)

var (
	_    benchmark.MsgServer = &Keeper{}
	once sync.Once
)

type Keeper struct {
	kvServiceMap     KVServiceMap
	telemetryService telemetry.Service
	validate         bool
	errExit          bool

	metricOpKey    []string
	metricTotalKey []string
	metricMissKey  []string
}

func NewKeeper(kvMap KVServiceMap, telemetryService telemetry.Service) *Keeper {
	k := &Keeper{
		kvServiceMap:     kvMap,
		telemetryService: telemetryService,
		validate:         false,
		errExit:          false,
		metricOpKey:      []string{"benchmark", "op"},
		metricTotalKey:   []string{"benchmark", "total"},
		metricMissKey:    []string{"benchmark", "miss"},
	}
	once.Do(func() {
		telemetryService.RegisterMeasure(k.metricOpKey, "op")
		telemetryService.RegisterCounter(k.metricMissKey, "op")
		telemetryService.RegisterCounter(k.metricTotalKey)
	})
	return k
}

func (k *Keeper) LoadTest(ctx context.Context, msg *benchmark.MsgLoadTest) (*benchmark.MsgLoadTestResponse, error) {
	res := &benchmark.MsgLoadTestResponse{}
	for _, op := range msg.Ops {
		k.telemetryService.IncrCounter(k.metricTotalKey, 1)
		err := k.executeOp(ctx, op)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func (k *Keeper) measureSince(since time.Time, opType string) {
	k.telemetryService.MeasureSince(since, k.metricOpKey, telemetry.Label{Name: "op", Value: opType})
}

func (k *Keeper) countMiss(opType string) {
	k.telemetryService.IncrCounter(k.metricMissKey, 1, telemetry.Label{Name: "op", Value: opType})
}

func (k *Keeper) executeOp(ctx context.Context, op *benchmark.Op) error {
	svc, ok := k.kvServiceMap[op.Actor]
	key := gen.Bytes(op.Seed, op.KeyLength)
	if !ok {
		return fmt.Errorf("actor %s not found", op.Actor)
	}
	start := time.Now()
	kv := svc.OpenKVStore(ctx)
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
