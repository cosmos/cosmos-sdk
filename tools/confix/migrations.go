package confix

import (
	"context"
	"fmt"
	"strings"

	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
)

const (
	AppConfig    = "app.toml"
	ClientConfig = "client.toml"
	CMTConfig    = "config.toml"
)

// MigrationMap defines a mapping from a version to a transformation plan.
type MigrationMap map[string]func(from *tomledit.Document, to string) transform.Plan

// loadDestConfigFile is the function signature to load the destination version
// configuration toml file.
type loadDestConfigFile func(to, planType string) (*tomledit.Document, error)

var Migrations = MigrationMap{
	"v0.45": NoPlan, // Confix supports only the current supported SDK version. So we do not support v0.44 -> v0.45.
<<<<<<< HEAD
	"v0.46": PlanBuilder,
	"v0.47": PlanBuilder,
	"v0.50": PlanBuilder,
	// "v0.xx.x": PlanBuilder, // add specific migration in case of configuration changes in minor versions
}

// PlanBuilder is a function that returns a transformation plan for a given diff between two files.
func PlanBuilder(from *tomledit.Document, to string) transform.Plan {
	plan := transform.Plan{}
	deletedSections := map[string]bool{}

	target, err := LoadLocalConfig(to)
=======
	"v0.46": defaultPlanBuilder,
	"v0.47": defaultPlanBuilder,
	"v0.50": defaultPlanBuilder,
	"v0.52": defaultPlanBuilder,
	"v2":    V2PlanBuilder,
	// "v0.xx.x": defaultPlanBuilder, // add specific migration in case of configuration changes in minor versions
}

type v2KeyChangesMap map[string][]string

// list all the keys which are need to be modified in v2
var v2KeyChanges = v2KeyChangesMap{
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
}

func defaultPlanBuilder(from *tomledit.Document, to, planType string) (transform.Plan, *tomledit.Document) {
	return PlanBuilder(from, to, planType, LoadLocalConfig)
}

// PlanBuilder is a function that returns a transformation plan for a given diff between two files.
func PlanBuilder(from *tomledit.Document, to, planType string, loadFn loadDestConfigFile) (transform.Plan, *tomledit.Document) {
	plan := transform.Plan{}
	deletedSections := map[string]bool{}

	target, err := loadFn(to, planType)
>>>>>>> 1d7f891ea (feat(confix): allow customization of migration plan (#21202))
	if err != nil {
		panic(fmt.Errorf("failed to parse file: %w. This file should have been valid", err))
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
			if diff.Type == Section {
				deletedSections[kv.Key] = true
				step = transform.Step{
					Desc: fmt.Sprintf("remove %s section", kv.Key),
					T:    transform.Remove(keys),
				}
			} else {
				// when the whole section is deleted we don't need to remove the keys
				if len(keys) > 1 && deletedSections[keys[0]] {
					continue
				}

				step = transform.Step{
					Desc: fmt.Sprintf("remove %s key", kv.Key),
					T:    transform.Remove(keys),
				}
			}
		}

		plan = append(plan, step)
	}

	return plan
}

// NoPlan returns a no-op plan.
func NoPlan(_ *tomledit.Document, to string) transform.Plan {
	fmt.Printf("no migration needed to %s\n", to)
	return transform.Plan{}
}
