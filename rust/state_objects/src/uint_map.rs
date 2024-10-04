use std::borrow::Borrow;
use ixc_core::{Context, Result};
use ixc_schema::state_object::{ObjectFieldValue, ObjectKey};
use crate::Map;

/// A map from keys to 128-bit unsigned integers.
pub struct UIntMap<K, V: UInt> {
    map: Map<K, V>,
}

pub trait UInt: ObjectFieldValue<In<'static>=Self, Out<'static>=Self> + Sized + Clone + 'static
{
    fn add(self, other: Self) -> Option<Self>;
    fn sub(self, other: Self) -> Option<Self>;
    const ZERO: Self;
}

impl UInt for u128 {
    fn add(self, other: Self) -> Option<Self> {
        self.checked_add(other)
    }

    fn sub(self, other: Self) -> Option<Self> {
        self.checked_sub(other)
    }

    const ZERO: Self = 0;
}

impl<K: ObjectKey, V: UInt> UIntMap<K, V> {
    /// Gets the current value for the given key, defaulting always to 0.
    pub fn get<'a, L>(&self, ctx: &Context, key: L) -> Result<V>
    where
        L: Borrow<K::In<'a>>,
    {
        // let value = self.map.get(ctx, key)?.clone();
        // Ok(value.unwrap_or(V::ZERO))
        todo!()
    }

    /// Adds the given value to the current value for the given key.
    pub fn add<'a, L>(&self, ctx: &mut Context, key: L, value: V) -> Result<V>
    where
        L: Borrow<K::In<'a>>,
    {
        // let current = self.get(ctx, key.borrow())?;
        // let new_value = V::add(current, value).ok_or_else(|| ())?;
        // self.map.set(ctx, key.borrow(), &new_value)?;
        // Ok(new_value)
        todo!()
    }

    /// Subtracts the given value from the current value for the given key,
    /// returning an error if the subtraction would result in a negative value.
    pub fn safe_sub(&self, ctx: &mut Context, key: &K::In<'_>, value: V) -> Result<V> {
        todo!()
    }
}
