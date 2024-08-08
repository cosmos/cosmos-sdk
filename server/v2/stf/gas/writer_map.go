package gas

import (
	"unsafe"

	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/store"
)

func NewMeteredWriterMap(conf gas.GasConfig, meter gas.Meter, state store.WriterMap) MeteredWriterMap {
	return MeteredWriterMap{
		config:             conf,
		meter:              meter,
		state:              state,
		cacheMeteredStores: make(map[string]*Store),
	}
}

// MeteredWriterMap wraps store.Writer and returns a gas metered
// version of it. Since the gas meter is shared across different
// writers, the metered writers are memoized.
type MeteredWriterMap struct {
	config             gas.GasConfig
	meter              gas.Meter
	state              store.WriterMap
	cacheMeteredStores map[string]*Store
}

func (m MeteredWriterMap) GetReader(actor []byte) (store.Reader, error) { return m.GetWriter(actor) }

func (m MeteredWriterMap) GetWriter(actor []byte) (store.Writer, error) {
	cached, ok := m.cacheMeteredStores[unsafeString(actor)]
	if ok {
		return cached, nil
	}

	state, err := m.state.GetWriter(actor)
	if err != nil {
		return nil, err
	}

	meteredState := NewStore(m.config, m.meter, state)
	m.cacheMeteredStores[string(actor)] = meteredState

	return meteredState, nil
}

func (m MeteredWriterMap) ApplyStateChanges(stateChanges []store.StateChanges) error {
	return m.state.ApplyStateChanges(stateChanges)
}

func (m MeteredWriterMap) GetStateChanges() ([]store.StateChanges, error) {
	return m.state.GetStateChanges()
}

func unsafeString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }
