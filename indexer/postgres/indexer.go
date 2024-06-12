package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"

	indexerbase "cosmossdk.io/indexer/base"
)

type indexer struct {
	ctx  context.Context
	conn *pgx.Conn
}

type Options struct {
	ConnectionURL string
}

func NewIndexer(ctx context.Context, opts Options) (indexerbase.Listener, error) {
	// get DATABASE_URL from environment
	dbUrl := opts.ConnectionURL
	if dbUrl == "" {
		var ok bool
		dbUrl, ok = os.LookupEnv("DATABASE_URL")
		if !ok {
			return indexerbase.Listener{}, fmt.Errorf("connection URL not set")
		}
	}

	conn, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		return indexerbase.Listener{}, err
	}

	i := &indexer{
		ctx:  ctx,
		conn: conn,
	}

	return i.logicalListener()
}

func (i *indexer) logicalListener() (indexerbase.Listener, error) {
	return indexerbase.Listener{
		Initialize:             i.initialize,
		InitializeModuleSchema: i.initModuleSchema,
		StartBlock:             i.startBlock,
		Commit:                 i.commit,
	}, nil
}

func (i *indexer) initialize(indexerbase.InitializationData) (int64, error) {
	// we don't care about persisting block data yet so just return 0
	return 0, nil
}

func (i *indexer) initModuleSchema(moduleName string, schema indexerbase.ModuleSchema) (int64, error) {
	//for _, _ := range schema.Tables {
	//}
	return -1, nil
}

func (i *indexer) startBlock(u uint64) error {
	return nil
}

func (i *indexer) commit() error {
	return nil
}
