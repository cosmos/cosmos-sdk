//! A u128 accumulator map.
use std::borrow::Borrow;
use num_enum::{IntoPrimitive, TryFromPrimitive};
use ixc_core::{fmt_error, Context, Result};
use ixc_core::error::{convert_client_error, convert_error_code, ClientError};
use ixc_core::resource::{InitializationError, StateObjectResource};
use ixc_core::result::ClientResult;
use ixc_message_api::code::ErrorCode;
use ixc_schema::state_object::ObjectKey;
use crate::Map;

// pub struct Accumulator {}

/// A map from keys to 128-bit unsigned integers that act as accumulators.
pub struct AccumulatorMap<K> {
    map: Map<K, u128>,
}


/// An error that can occur when performing a safe subtraction.
#[derive(Debug, Clone, TryFromPrimitive, IntoPrimitive)]
#[repr(u8)]
pub enum SafeSubError {
    /// The subtraction would result in a negative value.
    Underflow,
}

impl<K: ObjectKey> AccumulatorMap<K> {
    /// Gets the current value for the given key, defaulting always to 0.
    pub fn get<'a, L>(&self, ctx: &Context, key: L) -> ClientResult<u128>
    where
        L: Borrow<K::In<'a>>,
    {
        let value = self.map.get(ctx, key)?;
        Ok(value.unwrap_or_default())
    }

    /// Adds the given value to the current value for the given key.
    pub fn add<'a, L>(&self, ctx: &mut Context, key: L, value: u128) -> ClientResult<u128>
    where
        L: Borrow<K::In<'a>>,
    {
        let current = self.get(ctx, key.borrow())?;
        let new_value = current.saturating_add(value);
        self.map.set(ctx, key.borrow(), &new_value)?;
        Ok(new_value)
    }

    /// Subtracts the given value from the current value for the given key,
    /// returning an error if the subtraction would result in a negative value.
    pub fn safe_sub<'a, L>(&self, ctx: &mut Context, key: L, value: u128) -> ClientResult<u128, SafeSubError>
    where
        L: Borrow<K::In<'a>>,
    {
        let current = self.get(ctx, key.borrow()).map_err(convert_client_error)?;
        let new_value = current.checked_sub(value).ok_or_else(
            || ClientError::new(ErrorCode::HandlerCode(SafeSubError::Underflow), "".to_string())
        )?;
        self.map.set(ctx, key.borrow(), &new_value).map_err(convert_client_error)?;
        Ok(new_value)
    }
}

unsafe impl<K> StateObjectResource for AccumulatorMap<K> {
    unsafe fn new(scope: &[u8], prefix: u8) -> std::result::Result<Self, InitializationError> {
        Ok(AccumulatorMap {
            map: Map::new(scope, prefix)?,
        })
    }
}