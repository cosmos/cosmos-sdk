use interchain_core::{Context, Response};
use crate::codec::{ObjectKey, PrefixKey};

/// An index on a set of fields in a map which may map multiple index key values to a single primary key value.
pub struct Index<IndexKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(IndexKey, PrimaryKey)>,
}

impl<'a, IndexKey: ObjectKey<'a>, PrimaryKey: ObjectKey<'a>> Index<IndexKey, PrimaryKey> {
    /// Iterates over the index keys in the given range.
    pub fn iterate<Start, End>(&'a self, ctx: &Context, start: Start::Value, end: End::Value) -> Response<Iter<'a, IndexKey, PrimaryKey>>
    where
        Start: PrefixKey<'a, IndexKey>,
        End: PrefixKey<'a, IndexKey>,
    {
        todo!()
    }

    /// Iterates over the index keys in the given range in reverse order.
    pub fn iterate_reverse<Start, End>(&'a self, ctx: &Context, start: Start::Value, end: End::Value) -> Response<Iter<'a, IndexKey, PrimaryKey>>
    where
        Start: PrefixKey<'a, IndexKey>,
        End: PrefixKey<'a, IndexKey>,
    {
        todo!()
    }
}

/// An iterator over the index keys in a range.
pub struct Iter<'a, IndexKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(IndexKey, PrimaryKey)>,
    _phantom2: std::marker::PhantomData<&'a ()>,
}