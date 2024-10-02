//! The map module contains the `Map` struct, which represents a key-value map in storage.

use bump_scope::allocator_api2::alloc::Allocator;
use ixc_core::error::Error;
use ixc_core::{Context, Result};
use ixc_core::resource::{InitializationError, StateObject};
use ixc_core_macros::message_selector;
use ixc_message_api::handler::HandlerErrorCode;
use ixc_message_api::packet::MessagePacket;
use ixc_message_api::AccountID;
use ixc_message_api::header::MessageSelector;
use ixc_schema::state_object::{decode_object_value, encode_object_key, encode_object_value, ObjectKey, ObjectValue};

/// A key-value map.
pub struct Map<K, V> {
    _phantom: std::marker::PhantomData<(K, V)>,
    #[cfg(feature = "std")]
    prefix: Vec<u8>,
    // TODO no_std prefix
}

impl<K: ObjectKey, V: ObjectValue> Map<K, V> {
    // /// Checks if the map contains the given key.
    // pub fn has<'key>(&self, ctx: &Context<'key>, key: K::Value<'key>) -> Result<bool> {
    //     todo!()
    // }

    /// Gets the value of the map at the given key.
    pub fn get<'key, 'value>(&self, ctx: &'value Context<'key>, key: &K::In<'key>) -> Result<Option<V::Out<'value>>> {
        let key_bz = encode_object_key::<K, &dyn Allocator>(&self.prefix, key, ctx.memory_manager() as &dyn Allocator)
            .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;

        let value_bz = KVStoreClient.get(ctx, key_bz)?;
        let value_bz = match value_bz {
            None => return Ok(None),
            Some(value_bz) => value_bz,
        };

        let value = decode_object_value::<V>(value_bz, ctx.memory_manager()).
            map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        Ok(Some(value))
    }

    // /// Gets the value of the map at the given key, possibly from a previous block.
    // pub fn stale_get<'key>(&self, ctx: &Context<'key>, key: K::Value<'key>) -> Result<'key, V::Value<'key>> {
    //     todo!()
    // }

    /// Sets the value of the map at the given key.
    pub fn set<'key, 'value>(&self, ctx: &mut Context<'key>, key: &K::In<'key>, value: &V::In<'value>) -> Result<()> {
        let key_bz = encode_object_key::<K, &dyn Allocator>(&self.prefix, key, ctx.memory_manager() as &dyn Allocator)
            .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        let value_bz = encode_object_value::<V, &dyn Allocator>(value, ctx.memory_manager() as &dyn Allocator)
            .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        unsafe { KVStoreClient.set(ctx, key_bz, value_bz) }
    }

    // /// Updates the value of the map at the given key.
    // pub fn update<'key, 'value>(&self, ctx: &mut Context<'key>, key: K::In<'key>, updater: impl FnOnce(Option<V::In<'value>>) -> Option<V::In<'value>>) -> Result<()> {
    //     todo!()
    // }

    // /// Lazily updates the value of the map at the given key at some later point in time.
    // /// This function is unsafe because updater must be commutative and that cannot be guaranteed by the type system.
    // pub unsafe fn lazy_update<'a, 'b>(&self, ctx: &mut Context<'a>, key: K::Value<'a>, updater: impl FnOnce(Option<V::Value<'b>>) -> Option<V::Value<'b>>) -> Response<()> {
    //     todo!()
    // }
    //

    /// Deletes the value of the map at the given key.
    pub fn delete<'key>(&self, ctx: &mut Context<'key>, key: &K::In<'key>) -> Result<()> {
        let key_bz = encode_object_key::<K, &dyn Allocator>(&self.prefix, key, ctx.memory_manager() as &dyn Allocator)
            .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        unsafe { KVStoreClient.delete(ctx, key_bz) }
    }
}

const STATE_ACCOUNT: AccountID = AccountID::new(2);

const HAS_SELECTOR: MessageSelector = message_selector!("ixc.store.v1.has");
const GET_SELECTOR: MessageSelector = message_selector!("ixc.store.v1.get");
const SET_SELECTOR: MessageSelector = message_selector!("ixc.store.v1.set");
const DELETE_SELECTOR: MessageSelector = message_selector!("ixc.store.v1.delete");


fn create_packet<'a>(ctx: &'a Context, selector: MessageSelector) -> Result<&'a mut MessagePacket> {
    let mut packet = ctx.memory_manager().allocate_packet(0)
        .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
    let header = packet.header_mut();
    header.sender_account = ctx.account_id();
    header.account = STATE_ACCOUNT;
    header.message_selector = selector;
    Ok(packet)
}

struct KVStoreClient;

impl KVStoreClient {
    pub fn get<'a>(&self, ctx: &'a Context, key: &[u8]) -> Result<Option<&'a [u8]>> {
        let mut packet = create_packet(ctx, GET_SELECTOR)?;
        let header = packet.header_mut();
        unsafe {
            header.in_pointer1.set_slice(key);
            // TODO error code for not found
            ctx.host_backend().invoke(&mut packet, &ctx.memory_manager()).
                map_err(|_| todo!())?;
        }
        let res_bz = unsafe { packet.header().out_pointer1.get(packet) };
        Ok(Some(res_bz))
    }

    pub unsafe fn set(&self, ctx: &Context, key: &[u8], value: &[u8]) -> Result<()> {
        let mut packet = create_packet(ctx, SET_SELECTOR)?;
        let header = packet.header_mut();
        unsafe {
            header.in_pointer1.set_slice(key);
            header.in_pointer2.set_slice(value);
            ctx.host_backend().invoke(&mut packet, &ctx.memory_manager()).
                map_err(|_| todo!())?;
        }
        Ok(())
    }

    pub unsafe fn delete(&self, ctx: &Context, key: &[u8]) -> Result<()> {
        let mut packet = create_packet(ctx, DELETE_SELECTOR)?;
        let header = packet.header_mut();
        unsafe {
            header.in_pointer1.set_slice(key);
            ctx.host_backend().invoke(&mut packet, &ctx.memory_manager()).
                map_err(|_| todo!())?;
        }
        Ok(())
    }
}

unsafe impl <K, V> StateObject for Map<K, V> {
    unsafe fn new(scope: &[u8], p: u8) -> core::result::Result<Self, InitializationError> {
        let mut prefix = Vec::from(scope);
        prefix.push(p);
        Ok(Self {
            _phantom: std::marker::PhantomData,
            #[cfg(feature = "std")]
            prefix,
        })
    }
}