package main

import (
	"go/ast"

	migration "github.com/cosmos/cosmos-sdk/tools/migrate"
)

// callUpdates defines simple function argument truncation rules.
var callUpdates = []migration.FunctionArgUpdate{
	// No simple truncation changes in v53 -> v54.
}

// argSurgeries defines complex argument transformations using AST callbacks.
var argSurgeries = []migration.ArgSurgeryWithAST{
	// govkeeper.NewKeeper: removed StakingKeeper (pos 4), removed variadic InitOption,
	// added CalculateVoteResultsAndVotingPowerFn as last arg wrapping the removed StakingKeeper.
	//
	// v53: govkeeper.NewKeeper(cdc, storeService, acctKeeper, bankKeeper, stakingKeeper, distrKeeper, router, config, authority, ...initOptions)
	// v54: govkeeper.NewKeeper(cdc, storeService, acctKeeper, bankKeeper, distrKeeper, router, config, authority, govkeeper.NewDefaultCalculateVoteResultsAndVotingPower(stakingKeeper))
	{
		ImportPath:  "github.com/cosmos/cosmos-sdk/x/gov/keeper",
		FuncName:    "NewKeeper",
		OldArgCount: -1, // variadic, match any count >= 9
		Transform: func(pkgAlias string, args []ast.Expr) []ast.Expr {
			if len(args) < 9 {
				return args // unexpected arg count, leave unchanged
			}
			if hasDefaultGovVoteCalculator(args[len(args)-1], pkgAlias) {
				return args // already migrated
			}

			// args[4] is stakingKeeper — save it
			stakingKeeper := args[4]

			// Build new args: [0..3] + [5..8] (skip pos 4, skip any trailing variadic)
			newArgs := make([]ast.Expr, 0, 9)
			newArgs = append(newArgs, args[0:4]...) // cdc, storeService, acctKeeper, bankKeeper
			newArgs = append(newArgs, args[5:9]...) // distrKeeper, router, config, authority

			// Append: <pkgAlias>.NewDefaultCalculateVoteResultsAndVotingPower(stakingKeeper)
			newArgs = append(newArgs, &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   &ast.Ident{Name: pkgAlias},
					Sel: &ast.Ident{Name: "NewDefaultCalculateVoteResultsAndVotingPower"},
				},
				Args: []ast.Expr{stakingKeeper},
			})

			return newArgs
		},
	},
}

func hasDefaultGovVoteCalculator(expr ast.Expr, pkgAlias string) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return pkgIdent.Name == pkgAlias && sel.Sel.Name == "NewDefaultCalculateVoteResultsAndVotingPower"
}

// callArgEdits defines removal/addition of specific arguments from calls matched by pattern.
var callArgEdits = []migration.CallArgRemoval{
	// --- SetOrderEndBlockers: add bank at front ---
	{
		MethodName: "SetOrderEndBlockers",
		ArgsToAdd: []migration.ArgAddition{
			{Position: 0, Expr: "banktypes.ModuleName"},
		},
	},
}
