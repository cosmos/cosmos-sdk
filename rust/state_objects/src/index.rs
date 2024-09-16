use interchain_core::{Context, Response};
use crate::codec::{ObjectKey, PrefixKey};

pub struct Index<IndexKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(IndexKey, PrimaryKey)>,
}

impl<'a, IndexKey: ObjectKey, PrimaryKey: ObjectKey> Index<IndexKey, PrimaryKey> {
    pub fn iterate<Start, End>(&'a self, ctx: &Context, start: Start::Value<'a>, end: End::Value<'a>) -> Response<Iter<'a, IndexKey, PrimaryKey>>
    where
        Start: PrefixKey<IndexKey>,
        End: PrefixKey<IndexKey>,
    {
        todo!()
    }

    pub fn iterate_reverse<Start, End>(&'a self, ctx: &Context, start: Start::Value<'a>, end: End::Value<'a>) -> Response<Iter<'a, IndexKey, PrimaryKey>>
    where
        Start: PrefixKey<IndexKey>,
        End: PrefixKey<IndexKey>,
    {
        todo!()
    }
}

pub struct Iter<'a, IndexKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(IndexKey, PrimaryKey)>,
    _phantom2: std::marker::PhantomData<&'a ()>,
}