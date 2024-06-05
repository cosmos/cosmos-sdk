package codec

import indexerbase "cosmossdk.io/indexer/base"

type HasSchema interface {
	Fields() []indexerbase.Field
}
