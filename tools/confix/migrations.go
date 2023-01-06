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
		"v0.45": NoPlan,
		"v0.46": PlanBuilder,
		"v0.47": PlanBuilder,
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

	diffs := DiffDocs(from, target)
	for _, diff := range diffs {
		var step transform.Step
		if !diff.Deleted {
			step = transform.Step{
				Desc: fmt.Sprintf("migrate %s", diff.Key),
				T: transform.Func(func(_ context.Context, doc *tomledit.Document) error {
					_ = strings.Split(diff.Key, ".")

					if diff.Type == Section {

					} else if diff.Type == Mapping {

					} else {
						return fmt.Errorf("unknown diff type: %s", diff.Type)
					}

					return nil
				}),
			}
		} else {
			step = transform.Step{
				Desc: fmt.Sprintf("remove %s key", diff.Key),
				T:    transform.Remove(parser.Key{diff.Key}),
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
