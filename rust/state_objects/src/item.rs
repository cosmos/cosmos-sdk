//! The item module contains the `Item` struct, which represents a single item in storage.

use interchain_core::{Context};
use crate::Map;
use crate::response::Response;

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
