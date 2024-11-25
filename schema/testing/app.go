package schematesting

import (
	"fmt"

	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

// AppSchemaGen generates random valid app schemas, essentially a map of module names to module schemas.
var AppSchemaGen = rapid.Custom(func(t *rapid.T) map[string]schema.ModuleSchema {
	schema := make(map[string]schema.ModuleSchema)
	numModules := rapid.IntRange(1, 10).Draw(t, "numModules")
	for i := 0; i < numModules; i++ {
		moduleName := NameGen.Draw(t, "moduleName")
		moduleSchema := ModuleSchemaGen().Draw(t, fmt.Sprintf("moduleSchema[%s]", moduleName))
		schema[moduleName] = moduleSchema
	}
	return schema
})
