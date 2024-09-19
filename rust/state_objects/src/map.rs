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
    pub fn has<'key>(&self, ctx: &Context<'key>, key: K::Value<'key>) -> Response<bool> {
        todo!()
    }

    /// Gets the value of the map at the given key.
    pub fn get<'key, 'value>(&self, ctx: &Context<'key>, key: K::Value<'key>) -> Response<'value, V::Value<'value>> {
        todo!()
    }

    /// Gets the value of the map at the given key, possibly from a previous block.
    pub fn stale_get<'key>(&self, ctx: &Context<'key>, key: K::Value<'key>) -> Response<'key, V::Value<'key>> {
        todo!()
    }

    /// Sets the value of the map at the given key.
    pub fn set<'key, 'value>(&self, ctx: &mut Context<'key>, key: K::Value<'key>, value: V::Value<'value>) -> Response<()> {
        todo!()
    }

    /// Updates the value of the map at the given key.
    pub fn update<'key, 'value>(&self, ctx: &mut Context<'key>, key: K::Value<'key>, updater: impl FnOnce(Option<V::Value<'value>>) -> Option<V::Value<'value>>) -> Response<()> {
        todo!()
    }

    /// Lazily updates the value of the map at the given key at some later point in time.
    /// This function is unsafe because updater must be commutative and that cannot be guaranteed by the type system.
    pub unsafe fn lazy_update<'a, 'b>(&self, ctx: &mut Context<'a>, key: K::Value<'a>, updater: impl FnOnce(Option<V::Value<'b>>) -> Option<V::Value<'b>>) -> Response<()> {
        todo!()
    }

    /// Deletes the value of the map at the given key.
    pub fn delete<'key>(&self, ctx: &mut Context<'key>, key: K::Value<'key>) -> Response<()> {
        todo!()
    }
}

