use interchain_core::{Context, Response};
use crate::Item;

/// A sequence of unique, monotonically increasing 64-bit unsigned integers.
pub struct Seq {
    value: Item<u64>,
}

impl Seq {
    /// Peeks at the next value in the sequence without incrementing it.
    pub fn peek(&self, ctx: &Context) -> Response<u64> {
        todo!()
    }

    /// Increments the sequence and returns the new value.
    pub fn next(&self, ctx: &mut Context) -> Response<u64> {
        todo!()
    }

    /// Sets the sequence to the given value.
    /// This is unsafe and should only be used in exceptional circumstances.
    pub unsafe fn set(&self, ctx: &mut Context, value: u64) -> Response<()> {
        todo!()
    }
}