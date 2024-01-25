package gas

import (
	"unsafe"

	"cosmossdk.io/server/v2/core/stf"
	"cosmossdk.io/server/v2/core/store"
)

func NewMeteredWriterMap(conf StoreConfig, meter stf.GasMeter, state store.WriterMap) MeteredWriterMap {
	return MeteredWriterMap{
		config:             conf,
		meter:              meter,
		state:              state,
		cacheMeteredStores: make(map[string]*Store),
	}
}

type MeteredWriterMap struct {
	config             StoreConfig
	meter              stf.GasMeter
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

func newMeteredState(state store.WriterMap, meter stf.GasMeter) store.WriterMap {
	return MeteredWriterMap{
		meter:              meter,
		state:              state,
		cacheMeteredStores: make(map[string]*Store),
	}
}

func unsafeString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }
