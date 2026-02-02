package iavl

import (
	io "io"

	storetypes "cosmossdk.io/store/types"
)

type CommitStoreWrapper struct {
	tree *CommitTree
}

func (c CommitStoreWrapper) CacheWrap() storetypes.CacheWrap {
	//TODO implement me
	panic("implement me")
}

func (c CommitStoreWrapper) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	//TODO implement me
	panic("implement me")
}

//	func (c CommitStoreWrapper) Commit() storetypes.CommitID {
//		panic("CommitStoreWrapper does not support Commit()")
//	}
//
//	func (c CommitStoreWrapper) LastCommitID() storetypes.CommitID {
//		return c.tree.lastCommitId
//	}
//
//	func (c CommitStoreWrapper) WorkingHash() []byte {
//		return c.tree.lastCommitId.Hash
//	}
//
// func (c CommitStoreWrapper) SetPruning(options pruningtypes.PruningOptions) {}
//
//	func (c CommitStoreWrapper) GetPruning() pruningtypes.PruningOptions {
//		return pruningtypes.PruningOptions{}
//	}
//
//	func (c CommitStoreWrapper) GetStoreType() storetypes.StoreType {
//		return storetypes.StoreTypeIAVL
//	}
//
//	func (c CommitStoreWrapper) CacheWrap() storetypes.CacheWrap {
//		//TODO implement me
//		panic("implement me")
//	}
//
//	func (c CommitStoreWrapper) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
//		//TODO implement me
//		panic("implement me")
//	}
var _ storetypes.CacheWrapper = (*CommitStoreWrapper)(nil)
