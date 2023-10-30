package gas

import "cosmossdk.io/store/v2"

var _ store.Iterator = (*iterator)(nil)

type iterator struct {
	gasMeter  store.GasMeter
	gasConfig store.GasConfig
	parent    store.Iterator
}

func newIterator(parent store.Iterator, gm store.GasMeter, gc store.GasConfig) store.Iterator {
	return &iterator{
		parent:    parent,
		gasConfig: gc,
		gasMeter:  gm,
	}
}

func (itr *iterator) Domain() ([]byte, []byte) {
	return itr.parent.Domain()
}

func (itr *iterator) Valid() bool {
	return itr.parent.Valid()
}

func (itr *iterator) Key() []byte {
	return itr.parent.Key()
}

func (itr *iterator) Value() []byte {
	return itr.parent.Value()
}

func (itr *iterator) Next() bool {
	itr.consumeGasSeek()
	return itr.parent.Next()
}

func (itr *iterator) Close() {
	itr.parent.Close()
}

func (itr *iterator) Error() error {
	return itr.parent.Error()
}

// consumeGasSeek consumes a fixed amount of gas for each iteration step and a
// variable gas cost based on the current key and value's length. This is called
// prior to the iterator's Next() call.
func (itr *iterator) consumeGasSeek() {
	if itr.Valid() {
		key := itr.Key()
		value := itr.Value()

		itr.gasMeter.ConsumeGas(itr.gasConfig.ReadCostPerByte*store.Gas(len(key)), store.GasDescValuePerByte)
		itr.gasMeter.ConsumeGas(itr.gasConfig.ReadCostPerByte*store.Gas(len(value)), store.GasDescValuePerByte)
	}

	itr.gasMeter.ConsumeGas(itr.gasConfig.IterNextCostFlat, store.GasDescIterNextCostFlat)
}
