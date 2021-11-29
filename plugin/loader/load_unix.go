// +build cgo,!noplugin
// +build linux darwin freebsd

package loader

import (
	"errors"
	"plugin"

	cplugin "github.com/cosmos/cosmos-sdk/plugin"
)

func init() {
	loadPluginFunc = unixLoadPlugin
}

func unixLoadPlugin(fi string) ([]cplugin.Plugin, error) {
	pl, err := plugin.Open(fi)
	if err != nil {
		return nil, err
	}
	pls, err := pl.Lookup(cplugin.PLUGINS_SYMBOL)
	if err != nil {
		return nil, err
	}

	typePls, ok := pls.(*[]cplugin.Plugin)
	if !ok {
		return nil, errors.New("filed 'Plugins' didn't contain correct type")
	}

	return *typePls, nil
}
