package module

import (
	"context"
	"fmt"

	"cosmossdk.io/tools/benchmark"
	gen "cosmossdk.io/tools/benchmark/generator"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

var (
	_               benchmark.MsgServer = &Keeper{}
	metricOpKey                         = []string{"benchmark", "op"}
	metricGetKey                        = append(metricOpKey, "get")
	metricDelete                        = append(metricOpKey, "delete")
	metricInsertKey                     = append(metricOpKey, "insert")
	metricUpdateKey                     = append(metricOpKey, "update")
	metricTotalKey                      = []string{"benchmark", "total"}
	metricMissKey                       = []string{"benchmark", "miss"}
)

type Keeper struct {
	kvServiceMap KVServiceMap
	validate     bool
	errExit      bool
}

func NewKeeper(kvMap KVServiceMap) *Keeper {
	k := &Keeper{
		kvServiceMap: kvMap,
		validate:     false,
		errExit:      false,
	}
	return k
}

func (k *Keeper) LoadTest(ctx context.Context, msg *benchmark.MsgLoadTest) (*benchmark.MsgLoadTestResponse, error) {
	res := &benchmark.MsgLoadTestResponse{}
	for _, op := range msg.Ops {
		telemetry.IncrCounter(1, metricTotalKey...)
		err := k.executeOp(ctx, op)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func (k *Keeper) executeOp(ctx context.Context, op *benchmark.Op) error {
	svc, ok := k.kvServiceMap[op.Actor]
	key := gen.Bytes(op.Seed, op.KeyLength)
	if !ok {
		return fmt.Errorf("actor %s not found", op.Actor)
	}
	kv := svc.OpenKVStore(ctx)
	switch {
	case op.Delete:
		telemetry.IncrCounter(1, metricDelete...)
		if k.validate {
			exists, err := kv.Has(key)
			if err != nil {
				return err
			}
			if !exists {
				telemetry.IncrCounter(1, metricMissKey...)
				if k.errExit {
					return fmt.Errorf("key %d not found", op.Seed)
				}
			}
		}
		return kv.Delete(key)
	case op.ValueLength > 0:
		metricKey := metricInsertKey
		if op.Exists {
			metricKey = metricUpdateKey
		}
		telemetry.IncrCounter(1, metricKey...)
		if k.validate {
			exists, err := kv.Has(key)
			if err != nil {
				return err
			}
			if exists != op.Exists {
				telemetry.IncrCounter(1, metricMissKey...)
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
		telemetry.IncrCounter(1, metricGetKey...)
		v, err := kv.Get(key)
		if v == nil {
			// always count a miss on GET since it requires no extra I/O
			telemetry.IncrCounter(1, metricMissKey...)
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
