//! The map module contains the `Map` struct, which represents a key-value map in storage.

use std::iter::Product;
use interchain_core::{Context, Response};
use crate::codec::{ObjectKey, ObjectValue, PrefixKey};

/// A key-value map.
pub struct Map<K, V> {
    _phantom: std::marker::PhantomData<(K, V)>,
}

impl<'a, K: ObjectKey<'a>, V: ObjectValue<'a>> Map<K, V> {
    /// Checks if the map contains the given key.
    pub fn has(&self, ctx: &'a Context, key: K::Value) -> Response<bool> {
        todo!()
    }

    /// Gets the value of the map at the given key.
    pub fn get(&self, ctx: &'a Context, key: K::Value) -> Response<Option<V::Value>> {
        todo!()
    }

    /// Sets the value of the map at the given key.
    pub fn set(&self, ctx: &'a mut Context, key: K::Value, value: V::Value) -> Response<()> {
        todo!()
    }

    /// Updates the value of the map at the given key.
    pub fn update(&self, ctx: &'a mut Context, key: K::Value, updater: impl FnOnce(Option<V::Value>) -> Option<V::Value>) -> Response<()> {
        todo!()
    }

    /// Deletes the value of the map at the given key.
    pub fn delete(&self, ctx: &'a mut Context, key: K::Value) -> Response<()> {
        todo!()
    }
}

