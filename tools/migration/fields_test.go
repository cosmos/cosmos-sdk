package migration

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)

func TestUpdateStructFieldRemovals(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		removals     []StructFieldRemoval
		wantModified bool
		wantContains []string
		wantMissing  []string
	}{
		{
			name: "remove single field from struct",
			input: `package main
type SimApp struct {
	BankKeeper    bankkeeper.Keeper
	CircuitKeeper circuitkeeper.Keeper
	GovKeeper     govkeeper.Keeper
}`,
			removals: []StructFieldRemoval{
				{StructName: "SimApp", FieldName: "CircuitKeeper"},
			},
			wantModified: true,
			wantContains: []string{"BankKeeper", "GovKeeper"},
			wantMissing:  []string{"CircuitKeeper"},
		},
		{
			name: "remove multiple fields",
			input: `package main
type SimApp struct {
	BankKeeper    bankkeeper.Keeper
	CircuitKeeper circuitkeeper.Keeper
	NFTKeeper     nftkeeper.Keeper
	GovKeeper     govkeeper.Keeper
}`,
			removals: []StructFieldRemoval{
				{StructName: "SimApp", FieldName: "CircuitKeeper"},
				{StructName: "SimApp", FieldName: "NFTKeeper"},
			},
			wantModified: true,
			wantContains: []string{"BankKeeper", "GovKeeper"},
			wantMissing:  []string{"CircuitKeeper", "NFTKeeper"},
		},
		{
			name: "no match for wrong struct name",
			input: `package main
type OtherApp struct {
	CircuitKeeper circuitkeeper.Keeper
}`,
			removals: []StructFieldRemoval{
				{StructName: "SimApp", FieldName: "CircuitKeeper"},
			},
			wantModified: false,
			wantContains: []string{"CircuitKeeper"},
		},
		{
			name: "no match for wrong field name",
			input: `package main
type SimApp struct {
	BankKeeper bankkeeper.Keeper
}`,
			removals: []StructFieldRemoval{
				{StructName: "SimApp", FieldName: "CircuitKeeper"},
			},
			wantModified: false,
			wantContains: []string{"BankKeeper"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.AllErrors)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			modified, err := updateStructFieldRemovals(node, tt.removals)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if modified != tt.wantModified {
				t.Errorf("modified = %v, want %v", modified, tt.wantModified)
			}

			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, node); err != nil {
				t.Fatalf("print error: %v", err)
			}
			output := buf.String()

			for _, s := range tt.wantContains {
				if !strings.Contains(output, s) {
					t.Errorf("output should contain %q, got:\n%s", s, output)
				}
			}
			for _, s := range tt.wantMissing {
				if strings.Contains(output, s) {
					t.Errorf("output should NOT contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}

func TestUpdateStructFieldModifications(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		mods         []StructFieldModification
		wantModified bool
		wantContains []string
	}{
		{
			name: "make field a pointer",
			input: `package main
type SimApp struct {
	EpochsKeeper epochskeeper.Keeper
	BankKeeper   bankkeeper.Keeper
}`,
			mods: []StructFieldModification{
				{StructName: "SimApp", FieldName: "EpochsKeeper", MakePointer: true},
			},
			wantModified: true,
			wantContains: []string{"*epochskeeper.Keeper"},
		},
		{
			name: "already a pointer — no change",
			input: `package main
type SimApp struct {
	EpochsKeeper *epochskeeper.Keeper
}`,
			mods: []StructFieldModification{
				{StructName: "SimApp", FieldName: "EpochsKeeper", MakePointer: true},
			},
			wantModified: false,
			wantContains: []string{"*epochskeeper.Keeper"},
		},
		{
			name: "no match for wrong struct name",
			input: `package main
type Other struct {
	EpochsKeeper epochskeeper.Keeper
}`,
			mods: []StructFieldModification{
				{StructName: "SimApp", FieldName: "EpochsKeeper", MakePointer: true},
			},
			wantModified: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.AllErrors)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			modified, err := updateStructFieldModifications(node, tt.mods)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if modified != tt.wantModified {
				t.Errorf("modified = %v, want %v", modified, tt.wantModified)
			}

			var buf bytes.Buffer
			if err := printer.Fprint(&buf, fset, node); err != nil {
				t.Fatalf("print error: %v", err)
			}
			output := buf.String()

			for _, s := range tt.wantContains {
				if !strings.Contains(output, s) {
					t.Errorf("output should contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}
