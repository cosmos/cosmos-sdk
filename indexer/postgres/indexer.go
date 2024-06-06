package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"

	indexerbase "cosmossdk.io/indexer/base"
)

type indexer struct {
	conn *pgx.Conn
}

type Options struct{}

func NewIndexer(ctx context.Context, opts Options) (indexerbase.LogicalListener, error) {
	// get env var DATABASE_URL
	dbUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		panic("DATABASE_URL not set")
	}

	conn, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		panic(err)
	}

	i := &indexer{
		conn: conn,
	}
	return i.logicalListener()
}

func (i *indexer) logicalListener() (indexerbase.LogicalListener, error) {
	return indexerbase.LogicalListener{
		PhysicalListener: indexerbase.PhysicalListener{},
		EnsureSetup:      i.ensureSetup,
	}, nil
}

func (i *indexer) ensureSetup(data indexerbase.LogicalSetupData) error {
	for _, table := range data.Schema.Tables {
		createTable, err := i.createTableStatement(table)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", createTable)
	}
	return nil
}

//func (i indexer) StartBlock(u uint64) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (i indexer) IndexBlockHeader(data *indexerbase.BlockHeaderData) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (i indexer) IndexTx(data *indexerbase.TxData) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (i indexer) IndexEvent(data *indexerbase.EventData) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (i indexer) IndexEntityUpdate(update indexerbase.EntityUpdate) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (i indexer) IndexEntityDelete(entityDelete indexerbase.EntityDelete) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (i indexer) Commit() error {
//	//TODO implement me
//	panic("implement me")
//}
//
//var _ indexerbase.Indexer = &indexer{}
