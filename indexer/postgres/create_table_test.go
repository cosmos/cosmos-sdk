package postgres

import (
	"os"

	"cosmossdk.io/indexer/postgres/internal/testdata"
	"cosmossdk.io/schema"
)

func ExampleCreateTable_AllKinds() {
	exampleCreateTable(testdata.AllKindsObject)
	// Output:
	// CREATE TABLE IF NOT EXISTS "test_all_kinds" ("id" BIGINT NOT NULL,
	//	"string" TEXT NOT NULL,
	//	"bytes" BYTEA NOT NULL,
	//	"int8" SMALLINT NOT NULL,
	//	"uint8" SMALLINT NOT NULL,
	//	"int16" SMALLINT NOT NULL,
	//	"uint16" INTEGER NOT NULL,
	//	"int32" INTEGER NOT NULL,
	//	"uint32" BIGINT NOT NULL,
	//	"int64" BIGINT NOT NULL,
	//	"uint64" NUMERIC NOT NULL,
	//	"integer" NUMERIC NOT NULL,
	//	"decimal" NUMERIC NOT NULL,
	//	"bool" BOOLEAN NOT NULL,
	//	"time" TIMESTAMPTZ GENERATED ALWAYS AS (nanos_to_timestamptz("time_nanos")) STORED,
	//	"time_nanos" BIGINT NOT NULL,
	//	"duration" BIGINT NOT NULL,
	//	"float32" REAL NOT NULL,
	//	"float64" DOUBLE PRECISION NOT NULL,
	//	"bech32address" TEXT NOT NULL,
	//	"enum" "test_my_enum" NOT NULL,
	//	"json" JSONB NOT NULL,
	//	PRIMARY KEY ("id")
	// );
	// GRANT SELECT ON TABLE "test_all_kinds" TO PUBLIC;
}

func ExampleCreateTable_Singleton() {
	exampleCreateTable(testdata.SingletonObject)
	// Output:
	// CREATE TABLE IF NOT EXISTS "test_singleton" (_id INTEGER NOT NULL CHECK (_id = 1),
	//	"foo" TEXT NOT NULL,
	//	"bar" INTEGER NOT NULL,
	//	PRIMARY KEY (_id)
	// );
	// GRANT SELECT ON TABLE "test_singleton" TO PUBLIC;
}

func exampleCreateTable(objectType schema.ObjectType) {
	tm := NewTableManager("test", objectType, ManagerOptions{Logger: func(msg string, sql string, params ...interface{}) {}})
	err := tm.CreateTableSql(os.Stdout)
	if err != nil {
		panic(err)
	}
}
