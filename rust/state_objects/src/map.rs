//! The map module contains the `Map` struct, which represents a key-value map in storage.

use bump_scope::allocator_api2::alloc::Allocator;
use ixc_core::{Context, Result};
use ixc_core::error::Error;
use ixc_core_macros::message_selector;
use ixc_message_api::AccountID;
use ixc_message_api::handler::HandlerErrorCode;
use ixc_schema::binary::NativeBinaryCodec;
use ixc_schema::state_object::{ObjectKey, ObjectValue, PrefixKey};

/// A key-value map.
pub struct Map<K, V> {
    _phantom: std::marker::PhantomData<(K, V)>,
}

impl<K: ObjectKey, V: ObjectValue> Map<K, V> {
    // /// Checks if the map contains the given key.
    // pub fn has<'key>(&self, ctx: &Context<'key>, key: K::Value<'key>) -> Result<bool> {
    //     todo!()
    // }

    /// Gets the value of the map at the given key.
    pub fn get<'key, 'value>(&self, ctx: &Context<'key>, key: K::In<'key>) -> Result<'value, Option<V::In<'value>>> {
        let buf = K::encode(key, &ctx.memory_manager() as &dyn Allocator)
            .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        let mut packet = ctx.memory_manager().allocate_packet(0)
            .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        let header = packet.header_mut();
        header.sender_account = ctx.account_id();
        header.account = STATE_ACCOUNT;
        header.message_selector = GET_SELECTOR;
        unsafe {
            header.in_pointer1.set_slice(buf);
            // TODO error code for not found
            ctx.host_backend().invoke(&mut packet, &ctx.memory_manager()).
                map_err(|_| todo!())?;
        }
        NativeBinaryCodec::decode_object_key
        // decode result
        todo!()
    }

    // /// Gets the value of the map at the given key, possibly from a previous block.
    // pub fn stale_get<'key>(&self, ctx: &Context<'key>, key: K::Value<'key>) -> Result<'key, V::Value<'key>> {
    //     todo!()
    // }
    //
    /// Sets the value of the map at the given key.
    pub fn set<'key, 'value>(&self, ctx: &mut Context<'key>, key: K::In<'key>, value: V::In<'value>) -> Result<()> {
        todo!()
    }

    /// Updates the value of the map at the given key.
    pub fn update<'key, 'value>(&self, ctx: &mut Context<'key>, key: K::In<'key>, updater: impl FnOnce(Option<V::In<'value>>) -> Option<V::In<'value>>) -> Result<()> {
        todo!()
    }

    // /// Lazily updates the value of the map at the given key at some later point in time.
    // /// This function is unsafe because updater must be commutative and that cannot be guaranteed by the type system.
    // pub unsafe fn lazy_update<'a, 'b>(&self, ctx: &mut Context<'a>, key: K::Value<'a>, updater: impl FnOnce(Option<V::Value<'b>>) -> Option<V::Value<'b>>) -> Response<()> {
    //     todo!()
    // }
    //
    /// Deletes the value of the map at the given key.
    pub fn delete<'key>(&self, ctx: &mut Context<'key>, key: K::In<'key>) -> Result<()> {
        todo!()
    }
}

const STATE_ACCOUNT: AccountID = AccountID::new(2);

const HAS_SELECTOR: u64 = message_selector!("ixc.store.v1.has");
const GET_SELECTOR: u64 = message_selector!("ixc.store.v1.get");
const SET_SELECTOR: u64 = message_selector!("ixc.store.v1.set");
const DELETE_SELECTOR: u64 = message_selector!("ixc.store.v1.delete");
