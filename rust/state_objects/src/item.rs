//! The item module contains the `Item` struct, which represents a single item in storage.

use ixc_core::{Context, Response};
use crate::Map;

/// A single item in storage.
pub struct Item<V> {
    map: Map<(), V>
}

impl <V: Default> Item<V> {
    /// Gets the value of the item.
    pub fn get(&self, ctx: &Context) -> Response<V> {
        todo!()
    }

    /// Sets the value of the item.
    pub fn set(&self, ctx: &Context, value: V) -> Response<()> {
        todo!()
    }
}
