package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

// typeReplacements defines type/struct renames between v53 and v54.
var typeReplacements = []migration.TypeReplacement{
	// No simple type renames identified for v53 -> v54.
}

// fieldRemovals removes keeper fields from SimApp for deleted modules.
var fieldRemovals = []migration.StructFieldRemoval{}

// fieldModifications changes field types in SimApp.
var fieldModifications = []migration.StructFieldModification{
	// EpochsKeeper changed from value type to pointer in v54.
	{StructName: "SimApp", FieldName: "EpochsKeeper", MakePointer: true},
}
