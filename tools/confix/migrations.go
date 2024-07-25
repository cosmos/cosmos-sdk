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

type keyModificationMap map[string][]string

var keyModifications = map[string]keyModificationMap{
	"serverv2": {
		"min-retain-blocks": []string{"comet.min-retain-blocks"},
		"index-events":      []string{"comet.index-events"},
		"halt-height":       []string{"comet.halt-height"},
		"halt-time":         []string{"comet.halt-time"},
		"app-db-backend":    []string{"store.app-db-backend"},
		"pruning-keep-recent": []string{
			"store.options.ss-pruning-option.keep-recent",
			"store.options.sc-pruning-option.keep-recent",
		},
		"pruning-interval": []string{
			"store.options.ss-pruning-option.interval",
			"store.options.sc-pruning-option.interval",
		},
		"iavl-cache-size":       []string{"store.options.iavl-config.cache-size"},
		"iavl-disable-fastnode": []string{"store.options.iavl-config.skip-fast-storage-upgrade"},
		// Add other key mappings as needed
	},
	// "v0.xx.x": {}, // add keys to move for specific version if needed
}

// add extra steps if needed for specific version
var addditionalSteps = map[string]func(planType string) []transform.Step{
	"serverv2": func(planType string) []transform.Step {
		steps := []transform.Step{}
		if planType == "app" {
			step := transform.Step{
				Desc: "check and update app-db-backend value",
				T: transform.Func(func(_ context.Context, doc *tomledit.Document) error {
					// Get the value of the app-db-backend
					fieldKey := "store.app-db-backend"
					fieldKeys := strings.Split(fieldKey, ".")
					entry := doc.First(fieldKeys...)
					if entry == nil {
						return fmt.Errorf("no store.app-db-backend field found")
					}

					// check if app-db-backend value is empty and update it to goleveldb
					if entry.KeyValue != nil && entry.KeyValue.Value.String() == `""` {
						entry.KeyValue.Value = parser.MustValue("'goleveldb'")
					}

					return nil
				}),
			}
			steps = append(steps, step)
		}
		return steps
	},
	// "v0.xx.x": func(planType string) []transform.Step {},
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
	changes, ok := keyModifications[to]
	if ok {
		for oldKey, newKeys := range changes {
			oldKeysToModify = append(oldKeysToModify, oldKey)
			newKeysToModify = append(newKeysToModify, newKeys...)
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
			newKeys := changes[kv.Key]
			for _, newKey := range newKeys {
				if updatedKey, ok := keyUpdates[newKey]; ok {
					value := updatedKey.keyValue
					value.Value = parser.MustValue(kv.Value)
					plan = append(plan, transform.Step{
						Desc: fmt.Sprintf("add %s key", kv.Key),
						T:    transform.EnsureKey(updatedKey.tableKey, value),
					})
				}
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

	// check and run additional steps if found for "to" versions
	if stepsFunc, ok := addditionalSteps[to]; ok {
		plan = append(plan, stepsFunc(planType)...)
	}

	return plan
}

// NoPlan returns a no-op plan.
func NoPlan(_ *tomledit.Document, to, planType string) transform.Plan {
	fmt.Printf("no migration needed to %s\n", to)
	return transform.Plan{}
}
