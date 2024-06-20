package schemagen

import (
	"fmt"

	indexerbase "cosmossdk.io/schema"
	"pgregory.net/rapid"
)

var AppSchema = rapid.Custom(func(t *rapid.T) map[string]schema.ModuleSchema {
	schema := make(map[string]schema.ModuleSchema)
	numModules := rapid.IntRange(1, 10).Draw(t, "numModules")
	for i := 0; i < numModules; i++ {
		moduleName := Name.Draw(t, "moduleName")
		moduleSchema := ModuleSchema.Draw(t, fmt.Sprintf("moduleSchema[%s]", moduleName))
		schema[moduleName] = moduleSchema
	}
	return schema
})
