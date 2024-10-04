//! A u128 accumulator map.
use std::borrow::Borrow;
use ixc_core::{Context, Result};
use ixc_core::resource::{InitializationError, StateObjectResource};
use ixc_schema::state_object::ObjectKey;
use crate::Map;

// pub struct Accumulator {}

/// A map from keys to 128-bit unsigned integers that act as accumulators.
pub struct AccumulatorMap<K> {
    map: Map<K, u128>,
}

impl<K: ObjectKey> AccumulatorMap<K> {
    /// Gets the current value for the given key, defaulting always to 0.
    pub fn get<'a, L>(&self, ctx: &Context, key: L) -> Result<u128>
    where
        L: Borrow<K::In<'a>>,
    {
        let value = self.map.get(ctx, key)?;
        Ok(value.unwrap_or_default())
    }

    /// Adds the given value to the current value for the given key.
    pub fn add<'a, L>(&self, ctx: &mut Context, key: L, value: u128) -> Result<u128>
    where
        L: Borrow<K::In<'a>>,
    {
        let current = self.get(ctx, key.borrow())?;
        let new_value = current.checked_add(value).ok_or_else(|| ())?;
        self.map.set(ctx, key.borrow(), &new_value)?;
        Ok(new_value)
    }

    /// Subtracts the given value from the current value for the given key,
    /// returning an error if the subtraction would result in a negative value.
    pub fn safe_sub<'a, L>(&self, ctx: &mut Context, key: L, value: u128) -> Result<u128>
    where
        L: Borrow<K::In<'a>>,
    {
        let current = self.get(ctx, key.borrow())?;
        let new_value = current.checked_sub(value).ok_or_else(|| ())?;
        self.map.set(ctx, key.borrow(), &new_value)?;
        Ok(new_value)
    }
}

unsafe impl <K> StateObjectResource for AccumulatorMap<K> {
    unsafe fn new(scope: &[u8], prefix: u8) -> std::result::Result<Self, InitializationError> {
        Ok(AccumulatorMap {
            map: Map::new(scope, prefix)?,
        })
    }
}