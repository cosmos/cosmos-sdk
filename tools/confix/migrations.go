package confix

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
)

const (
	AppConfig        = "app.toml"
	AppConfigType    = "app"
	ClientConfig     = "client.toml"
	ClientConfigType = "client"
	CMTConfig        = "config.toml"
)

// MigrationMap defines a mapping from a version to a transformation plan.
type MigrationMap map[string]func(from *tomledit.Document, to, planType string) transform.Plan

var Migrations = MigrationMap{
	"v0.45":    NoPlan, // Confix supports only the current supported SDK version. So we do not support v0.44 -> v0.45.
	"v0.46":    PlanBuilder,
	"v0.47":    PlanBuilder,
	"v0.50":    PlanBuilder,
	"v0.52":    PlanBuilder,
	"serverv2": PlanBuilder,
	// "v0.xx.x": PlanBuilder, // add specific migration in case of configuration changes in minor versions
}

type KeyModificationMap map[string]string

var KeyModifications = map[string]KeyModificationMap{
	"serverv2": {
		"min-retain-blocks": "comet.min-retain-blocks",
		"index-events":      "comet.index-events",
		"halt-height":       "comet.halt-height",
		"halt-time":         "comet.min-retain-blocks",
		// Add other key mappings as needed
	},
	// "v0.xx.x": {}, // add keys to move for specific version if needed
}

type updatedKeyValue struct {
	tableKey parser.Key
	keyValue *parser.KeyValue
}

// PlanBuilder is a function that returns a transformation plan for a given diff between two files.
func PlanBuilder(from *tomledit.Document, to, planType string) transform.Plan {
	plan := transform.Plan{}
	deleteSections := []string{}
	deleteMappings := []Diff{}

	target, err := LoadLocalConfig(to, planType)
	if err != nil {
		panic(fmt.Errorf("failed to parse file: %w. This file should have been valid", err))
	}

	var oldKeysToModify, newKeysToModify []string
	keyUpdates := map[string]updatedKeyValue{}

	// check if key changes are needed with the "to" version
	changes, ok := KeyModifications[to]
	if ok {
		for oldKey, newKey := range changes {
			oldKeysToModify = append(oldKeysToModify, oldKey)
			newKeysToModify = append(newKeysToModify, newKey)
		}
	}

	diffs := DiffKeys(from, target)
	for _, diff := range diffs {
		diff := diff
		kv := diff.KV

		var step transform.Step
		keys := strings.Split(kv.Key, ".")

		if !diff.Deleted {
			switch diff.Type {
			case Section:
				step = transform.Step{
					Desc: fmt.Sprintf("add %s section", kv.Key),
					T: transform.Func(func(_ context.Context, doc *tomledit.Document) error {
						title := fmt.Sprintf("###                    %s Configuration                    ###", strings.Title(kv.Key))
						doc.Sections = append(doc.Sections, &tomledit.Section{
							Heading: &parser.Heading{
								Block: parser.Comments{
									strings.Repeat("#", len(title)),
									title,
									strings.Repeat("#", len(title)),
								},
								Name: keys,
							},
						})
						return nil
					}),
				}
			case Mapping:
				var tableKey parser.Key
				var keyValue *parser.KeyValue

				if len(keys) == 1 { // top-level key
					tableKey = nil
					keyValue = &parser.KeyValue{
						Block: kv.Block,
						Name:  parser.Key{keys[0]},
						Value: parser.MustValue(kv.Value),
					}
				} else if len(keys) > 1 {
					tableKey = keys[0 : len(keys)-1]
					keyValue = &parser.KeyValue{
						Block: kv.Block,
						Name:  parser.Key{keys[len(keys)-1]},
						Value: parser.MustValue(kv.Value),
					}
				} else {
					continue
				}

				if slices.Contains(newKeysToModify, kv.Key) {
					// store the key-value pair for later modification
					keyUpdates[kv.Key] = updatedKeyValue{tableKey, keyValue}
					continue
				} else {
					// create a step to add a new key-value pair
					step = transform.Step{
						Desc: fmt.Sprintf("add %s key", kv.Key),
						T:    transform.EnsureKey(tableKey, keyValue),
					}
				}

			default:
				panic(fmt.Errorf("unknown diff type: %s", diff.Type))
			}
		} else {
			// separate deleted mappings and sections for later processing
			if diff.Type == Mapping {
				deleteMappings = append(deleteMappings, diff)
			} else {
				deleteSections = append(deleteSections, kv.Key)
			}
			continue
		}

		plan = append(plan, step)
	}

	// process deleted mappings and update them if they are in the modifications list
	for _, mapping := range deleteMappings {
		kv := mapping.KV
		keys := strings.Split(kv.Key, ".")

		if slices.Contains(oldKeysToModify, kv.Key) {
			newKey := changes[kv.Key]
			if updatedKey, ok := keyUpdates[newKey]; ok {
				value := updatedKey.keyValue
				value.Value = parser.MustValue(kv.Value)
				plan = append(plan, transform.Step{
					Desc: fmt.Sprintf("add %s key", kv.Key),
					T:    transform.EnsureKey(updatedKey.tableKey, value),
				})
			}
		}

		// create a step to remove the old key-value pair
		step := transform.Step{
			Desc: fmt.Sprintf("remove %s key", kv.Key),
			T:    transform.Remove(keys),
		}
		plan = append(plan, step)
	}

	// create steps to remove old sections
	for _, section := range deleteSections {
		keys := strings.Split(section, ".")
		plan = append(plan, transform.Step{
			Desc: fmt.Sprintf("remove %s key", section),
			T:    transform.Remove(keys),
		})
	}

	return plan
}

// NoPlan returns a no-op plan.
func NoPlan(_ *tomledit.Document, to, planType string) transform.Plan {
	fmt.Printf("no migration needed to %s\n", to)
	return transform.Plan{}
}
