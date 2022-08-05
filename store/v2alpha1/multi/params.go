package multi

import (
	"fmt"
	"io"

	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"
)

// DefaultStoreParams returns a MultiStore config with an empty schema, a single backing DB,
// pruning with PruneDefault, no listeners and no tracer.
func DefaultStoreParams() StoreParams {
	return StoreParams{
		Pruning:          pruningtypes.NewPruningOptions(pruningtypes.PruningDefault),
		SchemaBuilder:    newSchemaBuilder(),
		storeKeys:        storeKeys{},
		traceListenMixin: newTraceListenMixin(),
	}
}

func (par *StoreParams) RegisterSubstore(skey types.StoreKey, typ types.StoreType) error {
	if !validSubStoreType(typ) {
		return fmt.Errorf("StoreType not supported: %v", typ)
	}
	var ok bool
	switch typ {
	case types.StoreTypePersistent:
		_, ok = skey.(*types.KVStoreKey)
	case types.StoreTypeMemory:
		_, ok = skey.(*types.MemoryStoreKey)
	case types.StoreTypeTransient:
		_, ok = skey.(*types.TransientStoreKey)
	}
	if !ok {
		return fmt.Errorf("invalid StoreKey for %v: %T", typ, skey)
	}
	if err := par.registerName(skey.Name(), typ); err != nil {
		return err
	}
	par.storeKeys[skey.Name()] = skey
	return nil
}

func (par *StoreParams) SetTracerFor(skey types.StoreKey, w io.Writer) {
	tlm := newTraceListenMixin()
	tlm.SetTracer(w)
	tlm.SetTracingContext(par.TraceContext)
	par.substoreTraceListenMixins[skey] = tlm
}

func (par *StoreParams) storeKey(key string) (types.StoreKey, error) {
	skey, ok := par.storeKeys[key]
	if !ok {
		return nil, fmt.Errorf("StoreKey instance not mapped: %s", key)
	}
	return skey, nil
}

func RegisterSubstoresFromMap[T types.StoreKey](par *StoreParams, keys map[string]T) error {
	for _, key := range keys {
		typ, err := types.StoreKeyToType(key)
		if err != nil {
			return err
		}
		if err = par.RegisterSubstore(key, typ); err != nil {
			return err
		}
	}
	return nil
}
