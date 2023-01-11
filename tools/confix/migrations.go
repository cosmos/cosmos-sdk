package confix

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
)

const (
	AppConfig    = "app.toml"
	ClientConfig = "client.toml"
	TMConfig     = "config.toml"
)

// MigrationMap defines a mapping from a version to a transformation plan.
type MigrationMap map[string]func(from *tomledit.Document, to string) transform.Plan

var (
	Migrations = MigrationMap{
		"v0.45": NoPlan, // Confix supports only the current supported SDK version. So we do not support v0.44 -> v0.45.
		"v0.46": PlanBuilder,
		"v0.47": PlanBuilder,
		// "v0.47.x": PlanBuilder, // add specific migration in case of configuration changes in minor versions
		// "v0.48": PlanBuilder,
	}

	//go:embed data
	data embed.FS
)

// PlanBuilder is a function that returns a transformation plan for a given diff between two files.
func PlanBuilder(from *tomledit.Document, to string) transform.Plan {
	plan := transform.Plan{}

	file, err := data.Open(fmt.Sprintf("data/%s-app.toml", to))
	if err != nil {
		panic(fmt.Errorf("failed to read file: %w. This file should have been included in confix", err))
	}

	target, err := tomledit.Parse(file)
	if err != nil {
		panic(fmt.Errorf("failed to parse file: %w. This file should have been valid", err))
	}

	deletedSections := map[string]bool{}

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
						doc.Sections = append(doc.Sections, &tomledit.Section{
							Heading: &parser.Heading{
								Block: parser.Comments{
									"###############################################################################",
									fmt.Sprintf("###							%s Configuration							###", strings.Title(kv.Key)),
									"###############################################################################"},
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
						})}
				} else if len(keys) > 1 {
					step = transform.Step{
						Desc: fmt.Sprintf("add %s key", kv.Key),
						T: transform.EnsureKey(parser.Key{keys[0]}, &parser.KeyValue{
							Block: kv.Block,
							Name:  parser.Key{keys[1]},
							Value: parser.MustValue(kv.Value),
						})}
				}
			default:
				panic(fmt.Errorf("unknown diff type: %s", diff.Type))
			}
		} else {
			if diff.Type == Section {
				deletedSections[kv.Key] = true
			}

			// when the whole section is deleted we don't need to remove the keys
			if len(keys) > 1 && deletedSections[keys[0]] {
				continue
			}

			step = transform.Step{
				Desc: fmt.Sprintf("remove %s key", kv.Key),
				T:    transform.Remove(keys),
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
