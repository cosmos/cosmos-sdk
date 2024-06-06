package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"

	indexerbase "cosmossdk.io/indexer/base"
)

type indexer struct {
	conn   *pgx.Conn
	tables map[string]*tableInfo
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
		conn:   conn,
		tables: map[string]*tableInfo{},
	}
	return i.logicalListener()
}

func (i *indexer) logicalListener() (indexerbase.LogicalListener, error) {
	return indexerbase.LogicalListener{
		PhysicalListener: indexerbase.PhysicalListener{
			StartBlock: i.startBlock,
			Commit:     i.commit,
		},
		EnsureSetup:    i.ensureSetup,
		OnEntityUpdate: i.onEntityUpdate,
	}, nil
}

func (i *indexer) ensureSetup(data indexerbase.LogicalSetupData) error {
	for _, table := range data.Schema.Tables {
		createTable, err := i.createTableStatement(table)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", createTable)
		i.tables[table.Name] = &tableInfo{table: table}
	}
	return nil
}

type tableInfo struct {
	table indexerbase.Table
}

func (i *indexer) startBlock(u uint64) error {
	return nil
}

//	func (i indexer) IndexBlockHeader(data *indexerbase.BlockHeaderData) error {
//		//TODO implement me
//		panic("implement me")
//	}
//
//	func (i indexer) IndexTx(data *indexerbase.TxData) error {
//		//TODO implement me
//		panic("implement me")
//	}
//
//	func (i indexer) IndexEvent(data *indexerbase.EventData) error {
//		//TODO implement me
//		panic("implement me")
//	}

func (i *indexer) onEntityUpdate(update indexerbase.EntityUpdate) error {
	ti, ok := i.tables[update.TableName]
	if !ok {
		return fmt.Errorf("table %s not found", update.TableName)
	}

	err := indexerbase.ValidateKey(ti.table.KeyColumns, update.Key)
	if err != nil {
		fmt.Printf("error validating key: %s\n", err)
	}

	if !update.Delete {
		err = indexerbase.ValidateValue(ti.table.ValueColumns, update.Value)
		if err != nil {
			fmt.Printf("error validating value: %s\n", err)
		}
	}

	return nil
}

func (i *indexer) commit() error {
	return nil
}
