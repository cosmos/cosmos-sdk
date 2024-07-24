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

type moveKeyMap map[string]struct {
	section string
	keyname string
}

var moveKeyMappings = map[string]moveKeyMap{
	"serverv2": {
		"min-retain-blocks": {
			section: "comet",
			keyname: "min-retain-blocks",
		},
		// Add other key mappings as needed
	},
	// "v0.xx.x": {}, // add keys to move for specific version if needed
}

// PlanBuilder is a function that returns a transformation plan for a given diff between two files.
func PlanBuilder(from *tomledit.Document, to, planType string) transform.Plan {
	plan := transform.Plan{}
	deleteSections := []string{}

	target, err := LoadLocalConfig(to, planType)
	if err != nil {
		panic(fmt.Errorf("failed to parse file: %w. This file should have been valid", err))
	}

	moveKeys := []string{}

	// check if key moves needed with "to" version
	moves, ok := moveKeyMappings[to]
	if ok {
		for oldKey, _ := range moves {
			moveKeys = append(moveKeys, oldKey)
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
				if len(keys) == 1 { // top-level key
					step = transform.Step{
						Desc: fmt.Sprintf("add %s key", kv.Key),
						T: transform.EnsureKey(nil, &parser.KeyValue{
							Block: kv.Block,
							Name:  parser.Key{keys[0]},
							Value: parser.MustValue(kv.Value),
						}),
					}
				} else if len(keys) > 1 {
					step = transform.Step{
						Desc: fmt.Sprintf("add %s key", kv.Key),
						T: transform.EnsureKey(keys[0:len(keys)-1], &parser.KeyValue{
							Block: kv.Block,
							Name:  parser.Key{keys[len(keys)-1]},
							Value: parser.MustValue(kv.Value),
						}),
					}
				}
			default:
				panic(fmt.Errorf("unknown diff type: %s", diff.Type))
			}
		} else {
			if diff.Type == Mapping {
				if slices.Contains(moveKeys, kv.Key) {
					newKey := moves[kv.Key]

					step = transform.Step{
						Desc: fmt.Sprintf("move %s key to %s key in section %s", kv.Key, newKey.keyname, newKey.section),
						T:    transform.MoveKey(keys, parser.Key{newKey.section}, parser.Key{newKey.keyname}),
					}
				} else {
					step = transform.Step{
						Desc: fmt.Sprintf("remove %s key", kv.Key),
						T:    transform.Remove(keys),
					}
				}
			} else {
				deleteSections = append(deleteSections, kv.Key)
				continue
			}
		}

		plan = append(plan, step)
	}

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
