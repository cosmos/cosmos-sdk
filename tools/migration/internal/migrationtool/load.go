package migrationtool

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var specOrder = []string{
	"group.yaml",
	"core.yaml",
	"crisis.yaml",
	"circuit.yaml",
	"nft.yaml",
	"gov.yaml",
	"epochs.yaml",
	"ante.yaml",
	"app-structure.yaml",
}

type Spec struct {
	ID           string       `yaml:"id"`
	Title        string       `yaml:"title"`
	Version      string       `yaml:"version"`
	Description  string       `yaml:"description"`
	Detection    Detection    `yaml:"detection"`
	Changes      Changes      `yaml:"changes"`
	ManualSteps  []ManualStep `yaml:"manual_steps"`
	Verification Verification `yaml:"verification"`
	SourcePath   string       `yaml:"-"`
}

type Detection struct {
	Imports  []string `yaml:"imports"`
	Patterns []string `yaml:"patterns"`
}

type Changes struct {
	GoMod             GoModChanges     `yaml:"go_mod"`
	Imports           ImportChanges    `yaml:"imports"`
	TextReplacements  []map[string]any `yaml:"text_replacements"`
	FileRemovals      []map[string]any `yaml:"file_removals"`
	StatementRemovals []map[string]any `yaml:"statement_removals"`
	MapEntryRemovals  []map[string]any `yaml:"map_entry_removals"`
	CallArgEdits      []map[string]any `yaml:"call_arg_edits"`
}

type GoModChanges struct {
	Update             map[string]string `yaml:"update"`
	Add                map[string]string `yaml:"add"`
	Remove             []string          `yaml:"remove"`
	StripLocalReplaces bool              `yaml:"strip_local_replaces"`
}

type ImportChanges struct {
	Rewrites []ImportRewrite `yaml:"rewrites"`
	Warnings []ImportWarning `yaml:"warnings"`
}

type ImportRewrite struct {
	Old string `yaml:"old"`
	New string `yaml:"new"`
}

type ImportWarning struct {
	Prefix  string `yaml:"prefix"`
	Message string `yaml:"message"`
	Fatal   bool   `yaml:"fatal"`
}

type ManualStep struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
}

type Verification struct {
	MustNotImport  []string      `yaml:"must_not_import"`
	MustNotContain PatternChecks `yaml:"must_not_contain"`
	MustContain    PatternChecks `yaml:"must_contain"`
}

type PatternChecks []PatternCheck

type PatternCheck struct {
	Pattern string
}

func (p *PatternCheck) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		p.Pattern = node.Value
		return nil
	case yaml.MappingNode:
		var raw struct {
			Pattern string `yaml:"pattern"`
		}
		if err := node.Decode(&raw); err != nil {
			return err
		}
		p.Pattern = raw.Pattern
		return nil
	default:
		return fmt.Errorf("unsupported pattern node kind %d", node.Kind)
	}
}

func LoadSpecs(specDir string) ([]Spec, error) {
	if stat, err := os.Stat(specDir); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", specDir)
	}

	var specs []Spec
	seen := make(map[string]bool)

	for _, name := range specOrder {
		fullPath := filepath.Join(specDir, name)
		spec, err := loadSpec(fullPath)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
		seen[name] = true
	}

	entries, err := os.ReadDir(specDir)
	if err != nil {
		return nil, err
	}

	var extras []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		if !seen[entry.Name()] {
			extras = append(extras, entry.Name())
		}
	}

	sort.Strings(extras)
	for _, name := range extras {
		spec, err := loadSpec(filepath.Join(specDir, name))
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}

	return specs, nil
}

func loadSpec(path string) (Spec, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, err
	}

	var spec Spec
	if err := yaml.Unmarshal(content, &spec); err != nil {
		return Spec{}, fmt.Errorf("parse %s: %w", path, err)
	}

	spec.SourcePath = path
	return spec, nil
}
