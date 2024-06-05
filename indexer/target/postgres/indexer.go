package postgres

import indexerbase "cosmossdk.io/indexer/base"

type indexer struct{}

func (i indexer) StartBlock(u uint64) error {
	//TODO implement me
	panic("implement me")
}

func (i indexer) MigrateSchema(data *indexerbase.MigrationData) error {
	//TODO implement me
	panic("implement me")
}

func (i indexer) IndexBlockHeader(data *indexerbase.BlockHeaderData) error {
	//TODO implement me
	panic("implement me")
}

func (i indexer) IndexTx(data *indexerbase.TxData) error {
	//TODO implement me
	panic("implement me")
}

func (i indexer) IndexEvent(data *indexerbase.EventData) error {
	//TODO implement me
	panic("implement me")
}

func (i indexer) IndexEntityUpdate(update indexerbase.EntityUpdate) error {
	//TODO implement me
	panic("implement me")
}

func (i indexer) IndexEntityDelete(entityDelete indexerbase.EntityDelete) error {
	//TODO implement me
	panic("implement me")
}

func (i indexer) Commit() error {
	//TODO implement me
	panic("implement me")
}

var _ indexerbase.Indexer = &indexer{}
