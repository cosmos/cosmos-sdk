package postgres

import (
	"os"

	"cosmossdk.io/indexer/postgres/internal/testdata"
	"cosmossdk.io/schema"
	"cosmossdk.io/schema/logutil"
)

func Example_objectIndexer_createTableSql_allKinds() {
	exampleCreateTable(testdata.AllKindsObject)
	// Output:
	// CREATE TABLE IF NOT EXISTS "test_all_kinds" (
	// 	"id" BIGINT NOT NULL,
	//	"ts" TIMESTAMPTZ GENERATED ALWAYS AS (nanos_to_timestamptz("ts_nanos")) STORED,
	//	"ts_nanos" BIGINT NOT NULL,
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
	//	"address" TEXT NOT NULL,
	//	"enum" "test_my_enum" NOT NULL,
	//	"json" JSONB NOT NULL,
	//	PRIMARY KEY ("id", "ts_nanos")
	// );
	// GRANT SELECT ON TABLE "test_all_kinds" TO PUBLIC;
}

func Example_objectIndexer_createTableSql_singleton() {
	exampleCreateTable(testdata.SingletonObject)
	// Output:
	// CREATE TABLE IF NOT EXISTS "test_singleton" (
	// 	_id INTEGER NOT NULL CHECK (_id = 1),
	//	"foo" TEXT NOT NULL,
	//	"bar" INTEGER NULL,
	//	"an_enum" "test_my_enum" NOT NULL,
	//	PRIMARY KEY (_id)
	// );
	// GRANT SELECT ON TABLE "test_singleton" TO PUBLIC;
}

func Example_objectIndexer_createTableSql_vote() {
	exampleCreateTable(testdata.VoteObject)
	// Output:
	// CREATE TABLE IF NOT EXISTS "test_vote" (
	// 	"proposal" BIGINT NOT NULL,
	// 	"address" TEXT NOT NULL,
	// 	"vote" "test_vote_type" NOT NULL,
	// 	_deleted BOOLEAN NOT NULL DEFAULT FALSE,
	// 	PRIMARY KEY ("proposal", "address")
	// );
	// GRANT SELECT ON TABLE "test_vote" TO PUBLIC;
}

func Example_objectIndexer_createTableSql_vote_no_retain_delete() {
	exampleCreateTableOpt(testdata.VoteObject, true)
	// Output:
	// CREATE TABLE IF NOT EXISTS "test_vote" (
	// 	"proposal" BIGINT NOT NULL,
	//	"address" TEXT NOT NULL,
	//	"vote" "test_vote_type" NOT NULL,
	//	PRIMARY KEY ("proposal", "address")
	// );
	// GRANT SELECT ON TABLE "test_vote" TO PUBLIC;
}

func exampleCreateTable(objectType schema.StateObjectType) {
	exampleCreateTableOpt(objectType, false)
}

func exampleCreateTableOpt(objectType schema.StateObjectType, noRetainDelete bool) {
	tm := newObjectIndexer("test", objectType, options{
		logger:                 logutil.NoopLogger{},
		disableRetainDeletions: noRetainDelete,
	})
	err := tm.createTableSql(os.Stdout)
	if err != nil {
		panic(err)
	}
}
