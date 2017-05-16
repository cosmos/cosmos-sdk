package main

import (
	"fmt"

	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/basecoin/plugins/counter"
)

type CounterPresenter struct{}

func (_ CounterPresenter) MakeKey(str string) ([]byte, error) {
	key := counter.New().StateKey()
	fmt.Println(string(key))
	return key, nil
}

func (_ CounterPresenter) ParseData(raw []byte) (interface{}, error) {
	fmt.Println("Data", len(raw))
	var cp counter.CounterPluginState
	err := wire.ReadBinaryBytes(raw, &cp)
	return cp, err
}
