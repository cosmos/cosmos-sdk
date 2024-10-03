//! The Message trait for invoking messages dynamically.

use allocator_api2::alloc::Allocator;
use ixc_message_api::handler::{HandlerError, HostBackend};
use ixc_message_api::header::MessageSelector;
use ixc_message_api::packet::MessagePacket;
use ixc_schema::codec::{decode_value, Codec};
use ixc_schema::mem::MemoryManager;
use ixc_schema::structs::StructSchema;
use ixc_schema::value::{OptionalValue, SchemaValue};
use crate::Context;

/// The Message trait for invoking messages dynamically.
pub trait Message<'a>: SchemaValue<'a> + StructSchema {
    /// The message selector.
    const SELECTOR: MessageSelector;
    /// The optional response type.
    type Response<'b>: OptionalValue<'b>;
    /// The optional error type.
    type Error: OptionalValue<'static>;
    /// The codec to use for encoding and decoding the message.
    type Codec: Codec + Default;
}

/// Extract the response and error types from a Result.
/// Used internally for building the Message trait with a macro.
pub trait ExtractResponseTypes {
    /// The response type.
    type Response;
    /// The error type.
    type Error;
}

impl<R, E> ExtractResponseTypes for core::result::Result<R, E> {
    type Response = R;
    type Error = E;
}

/// Handle a message packet with a message handler. Used for implementing macros.
pub unsafe fn handle_message<'c, 'b:'c, 'a: 'b, M: Message<'b>, H>(h: &H,
                                                            packet: &'a mut MessagePacket<'a>,
                                                            callbacks: &'a dyn HostBackend,
                                                            mem: &'a MemoryManager,
                                                            allocator: &'a dyn Allocator,
                                                            f: fn(&H, &'b mut Context<'b>, M)
                                                                  -> Result<<<M as Message<'b>>::Response<'c> as OptionalValue<'c>>::Value, ()>)
                                                            -> Result<(), HandlerError> {
    let cdc = M::Codec::default();
    let res = {
        let mut ctx = Context::new(packet, callbacks, &mem);
        let in1 = packet.header().in_pointer1.get(packet);
        let msg = decode_value::<M>(&cdc, in1, &mem).map_err(|e| HandlerError::Custom(0))?;
        f(h, &mut ctx, msg).map_err(|_| HandlerError::Custom(0))?
    };
    <M::Response<'_> as OptionalValue<'_>>::encode_value(&cdc, &res, packet, allocator).map_err(|e| HandlerError::Custom(0))
}

/// Handle a message packet with a message handler. Used for implementing macros.
pub unsafe fn handle_message2<'b, 'a: 'b, M: Message<'a>, H>(h: &H,
                                                             ctx: &'a mut Context<'a>,
                                                             allocator: &dyn Allocator,
                                                             f: fn(&H, &mut Context<'a>, M)
                                                                   -> Result<<<M as Message<'a>>::Response<'b> as OptionalValue<'b>>::Value, ()>)
                                                             -> Result<(), HandlerError> {
    // let cdc = M::Codec::default();
    // let in1 = ctx.message_packet.header().in_pointer1.get(ctx.message_packet);
    // let msg = decode_value::<M>(&cdc, in1, &ctx.mem).map_err(|e| HandlerError::Custom(0))?;
    // let res = f(h, ctx, msg).map_err(|_| HandlerError::Custom(0))?;
    // // <M::Response<'_> as OptionalValue<'_>>::encode_value(&cdc, &res, ctx.message_packet, allocator).map_err(|e| HandlerError::Custom(0))
    // // TODO encode response - need to mutably borrow the message packet to write the response
    // Ok(())
    todo!()
}
