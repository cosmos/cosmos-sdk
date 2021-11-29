// +build !cgo,!noplugin
// +build linux darwin freebsd

package loader

import (
	"errors"

	cplugin "github.com/cosmos/cosmos-sdk/plugin"
)

func init() {
	loadPluginFunc = nocgoLoadPlugin
}

func nocgoLoadPlugin(fi string) ([]cplugin.Plugin, error) {
	return nil, errors.New("not built with cgo support")
}
