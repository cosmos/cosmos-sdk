use interchain_core::{Context, Response};
use crate::codec::{ObjectKey, PrefixKey};

pub struct Index<IndexKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(IndexKey, PrimaryKey)>,
}

impl<'a, IndexKey: ObjectKey<'a>, PrimaryKey: ObjectKey<'a>> Index<IndexKey, PrimaryKey> {
    pub fn iterate<Start, End>(&'a self, ctx: &Context, start: Start::Value, end: End::Value) -> Response<Iter<'a, IndexKey, PrimaryKey>>
    where
        Start: PrefixKey<'a, IndexKey>,
        End: PrefixKey<'a, IndexKey>,
    {
        todo!()
    }

    pub fn iterate_reverse<Start, End>(&'a self, ctx: &Context, start: Start::Value, end: End::Value) -> Response<Iter<'a, IndexKey, PrimaryKey>>
    where
        Start: PrefixKey<'a, IndexKey>,
        End: PrefixKey<'a, IndexKey>,
    {
        todo!()
    }
}

pub struct Iter<'a, IndexKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(IndexKey, PrimaryKey)>,
    _phantom2: std::marker::PhantomData<&'a ()>,
}