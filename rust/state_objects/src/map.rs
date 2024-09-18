//! The map module contains the `Map` struct, which represents a key-value map in storage.

use std::iter::Product;
use interchain_core::{Context};
use interchain_schema::state_object::{ObjectKey, ObjectValue, PrefixKey};
use crate::response::Response;

/// A key-value map.
pub struct Map<K, V> {
    _phantom: std::marker::PhantomData<(K, V)>,
}

impl<K: ObjectKey, V: ObjectValue> Map<K, V> {
    /// Checks if the map contains the given key.
    pub fn has<'a>(&self, ctx: &Context<'a>, key: K::Value<'a>) -> Response<bool> {
        todo!()
    }

    /// Gets the value of the map at the given key.
    pub fn get<'a, 'b>(&self, ctx: &Context<'a>, key: K::Value<'a>) -> Response<'b, V::Value<'b>> {
        todo!()
    }

    /// Gets the value of the map at the given key, possibly from a previous block.
    pub fn stale_get<'a, 'b, 'c>(&self, ctx: &Context<'a>, key: K::Value<'b>) -> Response<'c, V::Value<'c>> {
        todo!()
    }

    /// Sets the value of the map at the given key.
    pub fn set<'a, 'b>(&self, ctx: &mut Context<'a>, key: K::Value<'a>, value: V::Value<'b>) -> Response<()> {
        todo!()
    }

    /// Updates the value of the map at the given key.
    pub fn update<'a, 'b>(&self, ctx: &mut Context<'a>, key: K::Value<'a>, updater: impl FnOnce(Option<V::Value<'b>>) -> Option<V::Value<'b>>) -> Response<()> {
        todo!()
    }

    /// Lazily updates the value of the map at the given key at some later point in time.
    /// This function is unsafe because updater must be commutative and that cannot be guaranteed by the type system.
    pub unsafe fn lazy_update<'a, 'b>(&self, ctx: &mut Context<'a>, key: K::Value<'a>, updater: impl FnOnce(Option<V::Value<'b>>) -> Option<V::Value<'b>>) -> Response<()> {
        todo!()
    }

    /// Deletes the value of the map at the given key.
    pub fn delete<'a>(&self, ctx: &mut Context<'a>, key: K::Value<'a>) -> Response<()> {
        todo!()
    }
}

