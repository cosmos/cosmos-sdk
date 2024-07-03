package postgres

import (
	"os"

	"cosmossdk.io/indexer/postgres/internal/testdata"
)

func ExampleCreateEnumType() {
	CreateEnumTypeSql(os.Stdout, "test", testdata.MyEnum)
	// Output:
	// CREATE TYPE "test_my_enum" AS ENUM ('a', 'b', 'c');
}
