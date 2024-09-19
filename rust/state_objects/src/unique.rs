use interchain_core::{Context, Response};

/// Enforces a queryable, uniqueness constraint on a set of fields in a map.
pub struct UniqueIndex<UniqueKey, PrimaryKey> {
    _phantom: std::marker::PhantomData<(UniqueKey, PrimaryKey)>,
}

impl<UniqueKey, PrimaryKey> UniqueIndex<UniqueKey, PrimaryKey> {
    /// Gets the primary key for the given unique key.
    pub fn get(&self, ctx: &Context, key: &UniqueKey) -> Response<Option<PrimaryKey>> {
        todo!()
    }
}