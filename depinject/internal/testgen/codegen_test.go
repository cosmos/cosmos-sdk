//go:build depinject

package testgen

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestBuild(t *testing.T) {
	_, _, _, _, err := Build(ModuleA{}, ModuleB{})
	assert.NilError(t, err)
}
