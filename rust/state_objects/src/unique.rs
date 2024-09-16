use interchain_core::{Context, Response};

pub struct UniqueIndex<IndexKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(IndexKey, PrimaryKey)>,
}

impl<IndexKey, PrimaryKey> UniqueIndex<IndexKey, PrimaryKey> {
    pub fn get(&self, ctx: &Context, key: &IndexKey) -> Response<Option<PrimaryKey>> {
        todo!()
    }
}