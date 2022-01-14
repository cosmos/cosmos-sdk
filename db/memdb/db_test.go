package memdb

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/dbtest"
)

func BenchmarkMemDBRangeScans1M(b *testing.B) {
	dbm := NewDB()
	defer dbm.Close()

	dbtest.BenchmarkRangeScans(b, dbm.ReadWriter(), int64(1e6))
}

func BenchmarkMemDBRangeScans10M(b *testing.B) {
	dbm := NewDB()
	defer dbm.Close()

	dbtest.BenchmarkRangeScans(b, dbm.ReadWriter(), int64(10e6))
}

func BenchmarkMemDBRandomReadsWrites(b *testing.B) {
	dbm := NewDB()
	defer dbm.Close()

	dbtest.BenchmarkRandomReadsWrites(b, dbm.ReadWriter())
}

func load(t *testing.T, _ string) db.DBConnection {
	return NewDB()
}

func TestGetSetHasDelete(t *testing.T) {
	dbtest.DoTestGetSetHasDelete(t, load)
}

func TestIterators(t *testing.T) {
	dbtest.DoTestIterators(t, load)
}

func TestVersioning(t *testing.T) {
	dbtest.DoTestVersioning(t, load)
}

func TestRevert(t *testing.T) {
	dbtest.DoTestRevert(t, load, false)
}

func TestTransactions(t *testing.T) {
	dbtest.DoTestTransactions(t, load, false)
}
