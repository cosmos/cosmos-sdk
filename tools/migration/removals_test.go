package migration

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"
)

func TestUpdateStatementRemovals(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		removals     []StatementRemoval
		wantModified bool
		wantContains []string
		wantMissing  []string
	}{
		{
			name: "remove assignment by target",
			input: `package main
func f() {
	app.Foo = newFoo()
	app.Bar = newBar()
}`,
			removals: []StatementRemoval{
				{AssignTarget: "app.Foo"},
			},
			wantModified: true,
			wantContains: []string{"app.Bar = newBar()"},
			wantMissing:  []string{"app.Foo"},
		},
		{
			name: "remove call by pattern",
			input: `package main
func f() {
	app.BaseApp.SetCircuitBreaker(x)
	app.Bar = newBar()
}`,
			removals: []StatementRemoval{
				{CallPattern: "app.BaseApp.SetCircuitBreaker"},
			},
			wantModified: true,
			wantContains: []string{"app.Bar = newBar()"},
			wantMissing:  []string{"SetCircuitBreaker"},
		},
		{
			name: "remove with IncludeFollowing",
			input: `package main
func f() {
	app.CircuitKeeper = circuitkeeper.NewKeeper(a, b)
	app.BaseApp.SetCircuitBreaker(x)
	app.Bar = newBar()
}`,
			removals: []StatementRemoval{
				{AssignTarget: "app.CircuitKeeper", IncludeFollowing: 1},
			},
			wantModified: true,
			wantContains: []string{"app.Bar = newBar()"},
			wantMissing:  []string{"CircuitKeeper", "SetCircuitBreaker"},
		},
		{
			name: "remove with IncludePrecedingAssign",
			input: `package main
func f() {
	groupConfig := group.DefaultConfig()
	app.GroupKeeper = groupkeeper.NewKeeper(groupConfig, a, b)
	app.Bar = newBar()
}`,
			removals: []StatementRemoval{
				{AssignTarget: "app.GroupKeeper", IncludePrecedingAssign: "groupConfig"},
			},
			wantModified: true,
			wantContains: []string{"app.Bar = newBar()"},
			wantMissing:  []string{"groupConfig", "GroupKeeper"},
		},
		{
			name: "no match leaves code unchanged",
			input: `package main
func f() {
	app.Bar = newBar()
}`,
			removals: []StatementRemoval{
				{AssignTarget: "app.Foo"},
			},
			wantModified: false,
			wantContains: []string{"app.Bar = newBar()"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.AllErrors)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			modified, err := updateStatementRemovals(node, tt.removals)
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

func TestUpdateCallArgRemovals(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		removals     []CallArgRemoval
		wantModified bool
		wantContains []string
		wantMissing  []string
	}{
		{
			name: "remove args from function call by FuncPattern",
			input: `package main
func f() {
	storetypes.NewKVStoreKeys(a.StoreKey, b.StoreKey, c.StoreKey, d.StoreKey)
}`,
			removals: []CallArgRemoval{
				{FuncPattern: "storetypes.NewKVStoreKeys", ArgsToRemove: []string{"b.StoreKey", "c.StoreKey"}},
			},
			wantModified: true,
			wantContains: []string{"a.StoreKey", "d.StoreKey"},
			wantMissing:  []string{"b.StoreKey", "c.StoreKey"},
		},
		{
			name: "remove args by MethodName",
			input: `package main
func f() {
	app.ModuleManager.SetOrderEndBlockers(a.ModuleName, b.ModuleName, c.ModuleName)
}`,
			removals: []CallArgRemoval{
				{MethodName: "SetOrderEndBlockers", ArgsToRemove: []string{"b.ModuleName"}},
			},
			wantModified: true,
			wantContains: []string{"a.ModuleName", "c.ModuleName"},
			wantMissing:  []string{"b.ModuleName"},
		},
		{
			name: "add arg at position 0",
			input: `package main
func f() {
	app.ModuleManager.SetOrderEndBlockers(a.ModuleName, c.ModuleName)
}`,
			removals: []CallArgRemoval{
				{
					MethodName: "SetOrderEndBlockers",
					ArgsToAdd: []ArgAddition{
						{Position: 0, Expr: "banktypes.ModuleName"},
					},
				},
			},
			wantModified: true,
			wantContains: []string{"banktypes.ModuleName, a.ModuleName"},
		},
		{
			name: "remove and add args simultaneously",
			input: `package main
func f() {
	app.ModuleManager.SetOrderEndBlockers(group.ModuleName, a.ModuleName)
}`,
			removals: []CallArgRemoval{
				{
					MethodName:   "SetOrderEndBlockers",
					ArgsToRemove: []string{"group.ModuleName"},
					ArgsToAdd:    []ArgAddition{{Position: 0, Expr: "banktypes.ModuleName"}},
				},
			},
			wantModified: true,
			wantContains: []string{"banktypes.ModuleName"},
			wantMissing:  []string{"group.ModuleName"},
		},
		{
			name: "skip duplicate arg additions",
			input: `package main
func f() {
	app.ModuleManager.SetOrderEndBlockers(banktypes.ModuleName, a.ModuleName)
}`,
			removals: []CallArgRemoval{
				{
					MethodName: "SetOrderEndBlockers",
					ArgsToAdd: []ArgAddition{
						{Position: 0, Expr: "banktypes.ModuleName"},
					},
				},
			},
			wantModified: false,
			wantContains: []string{"banktypes.ModuleName, a.ModuleName"},
		},
		{
			name: "no match leaves code unchanged",
			input: `package main
func f() {
	fmt.Println("hello")
}`,
			removals: []CallArgRemoval{
				{FuncPattern: "storetypes.NewKVStoreKeys", ArgsToRemove: []string{"x"}},
			},
			wantModified: false,
			wantContains: []string{`fmt.Println("hello")`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, "", tt.input, parser.AllErrors)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			modified, err := updateCallArgRemovals(node, tt.removals)
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

func TestUpdateMapEntryRemovals(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		removals     []MapEntryRemoval
		wantModified bool
		wantContains []string
		wantMissing  []string
	}{
		{
			name: "remove entry from var declaration map literal",
			input: `package main
var maccPerms = map[string][]string{
	a.ModuleName: nil,
	nft.ModuleName: nil,
	b.ModuleName: {"burner"},
}`,
			removals: []MapEntryRemoval{
				{MapVarName: "maccPerms", KeysToRemove: []string{"nft.ModuleName"}},
			},
			wantModified: true,
			wantContains: []string{"a.ModuleName", "b.ModuleName"},
			wantMissing:  []string{"nft.ModuleName"},
		},
		{
			name: "remove entry from assignment map literal",
			input: `package main
func f() {
	maccPerms = map[string][]string{
		a.ModuleName: nil,
		nft.ModuleName: nil,
	}
}`,
			removals: []MapEntryRemoval{
				{MapVarName: "maccPerms", KeysToRemove: []string{"nft.ModuleName"}},
			},
			wantModified: true,
			wantContains: []string{"a.ModuleName"},
			wantMissing:  []string{"nft.ModuleName"},
		},
		{
			name: "no match for different map var name",
			input: `package main
var otherMap = map[string]string{
	"key": "val",
}`,
			removals: []MapEntryRemoval{
				{MapVarName: "maccPerms", KeysToRemove: []string{`"key"`}},
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

			modified, err := updateMapEntryRemovals(node, tt.removals)
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

func TestParseExprSafe(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"banktypes.ModuleName", "banktypes.ModuleName"},
		{"x.Y.Z", "x.Y.Z"},
		{"42", "42"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parseExprSafe(tt.input)
			if err != nil {
				t.Fatalf("parseExprSafe(%q) error: %v", tt.input, err)
			}
			if expr == nil {
				t.Fatalf("parseExprSafe(%q) returned nil", tt.input)
			}
			got := exprToString(expr)
			if got != tt.want {
				t.Errorf("exprToString(parseExprSafe(%q)) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExprToString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"x", "x"},
		{"x.Y", "x.Y"},
		{"x.Y.Z", "x.Y.Z"},
		{"*x", "*x"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expr, err := parseExprSafe(tt.input)
			if err != nil {
				t.Fatalf("parseExprSafe error: %v", err)
			}
			got := exprToString(expr)
			if got != tt.want {
				t.Errorf("exprToString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
