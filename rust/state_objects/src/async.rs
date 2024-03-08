use dashu_int::UBig;
use cosmossdk_core::{BeginWriteContext, ReadContext, WriteContext};
use crate::{KeyCodec, ValueCodec};

struct UBigMap<K> {}

impl <K:KeyCodec> UBigMap<K> {
    fn read(&self, ctx: &ReadContext, key: K::T) -> Option<UBig> {
        todo!()
    }

    fn begin_read_write(&self, ctx: &ReadContext, key: K::T) -> cosmossdk_core::Result<ReadWriter<UBig>> {
        todo!()
    }

    fn begin_safe_sub(&self, ctx: &BeginWriteContext, key: K::T) -> cosmossdk_core::Result<ExecWriter<UBig>> {
        todo!()
    }

    fn begin_add(&self, ctx: &BeginWriteContext, key: K::T) -> cosmossdk_core::Result<ExecWriter<UBig>> {
        todo!()
    }
}

struct ExecWriter<V> {
}

impl <V:ValueCodec> ExecWriter<V> {
    fn exec(&self, ctx: &WriteContext, value: V::T) { }
}

struct ReadWriter<V> {
}

impl <V:ValueCodec> ReadWriter<V> {
    fn read(&self, ctx: &WriteContext) -> V::T {}
    fn write(&self, ctx: &WriteContext, value: V::T) {}
}
