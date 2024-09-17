//! The map module contains the `Map` struct, which represents a key-value map in storage.

use std::iter::Product;
use interchain_core::{Context, Response};
use interchain_schema::state_object::{ObjectKey, ObjectValue, PrefixKey};

/// A key-value map.
pub struct Map<K, V> {
    _phantom: std::marker::PhantomData<(K, V)>,
}

impl<K: ObjectKey, V: ObjectValue> Map<K, V> {
    /// Checks if the map contains the given key.
    pub fn has(&self, ctx: &Context, key: K::Value<'_>) -> Response<bool> {
        todo!()
    }

    /// Gets the value of the map at the given key.
    pub fn get<'a>(&self, ctx: &Context, key: K::Value<'_>) -> Response<'a, Option<V::Value<'a>>> {
        todo!()
    }

    /// Sets the value of the map at the given key.
    pub fn set(&self, ctx: &mut Context, key: K::Value<'_>, value: V::Value<'_>) -> Response<()> {
        todo!()
    }

    /// Updates the value of the map at the given key.
    pub fn update<'a>(&self, ctx: &'a mut Context, key: K::Value<'_>, updater: impl FnOnce(Option<V::Value<'a>>) -> Option<V::Value<'a>>) -> Response<()> {
        todo!()
    }

    /// Deletes the value of the map at the given key.
    pub fn delete(&self, ctx: &mut Context, key: K::Value<'_>) -> Response<()> {
        todo!()
    }
}

