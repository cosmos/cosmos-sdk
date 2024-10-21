use ixc_core::{Context, Result};
use ixc_schema::state_object::{ObjectKey, PrefixKey};

/// An index on a set of fields in a map which may map multiple index key values to a single primary key value.
pub struct Index<IndexKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(IndexKey, PrimaryKey)>,
}

impl<IndexKey: ObjectKey, PrimaryKey: ObjectKey> Index<IndexKey, PrimaryKey> {
    /// Iterates over the index keys in the given range.
    pub fn iterate<'a, Start, End>(&self, ctx: &'a Context, start: Start::Value<'a>, end: End::Value<'a>) -> Result<Iter<'a, IndexKey, PrimaryKey>>
    where
        Start: PrefixKey<IndexKey>,
        End: PrefixKey<IndexKey>,
    {
        todo!()
    }

    /// Iterates over the index keys in the given range in reverse order.
    pub fn iterate_reverse<'a, Start, End>(&self, ctx: &Context, start: Start::Value<'a>, end: End::Value<'a>) -> Result<Iter<'a, IndexKey, PrimaryKey>>
    where
        Start: PrefixKey<IndexKey>,
        End: PrefixKey<IndexKey>,
    {
        todo!()
    }
}

/// An iterator over the index keys in a range.
pub struct Iter<'a, IndexKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(IndexKey, PrimaryKey)>,
    _phantom2: std::marker::PhantomData<&'a ()>,
}