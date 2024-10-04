//! Low-level utilities for working with message structs and message packets directly.
use allocator_api2::alloc::Allocator;
use ixc_message_api::AccountID;
use ixc_message_api::handler::{HandlerErrorCode};
use ixc_message_api::packet::MessagePacket;
use ixc_schema::buffer::WriterFactory;
use ixc_schema::codec::{decode_value, Codec};
use ixc_schema::encoder::EncodeError;
use ixc_schema::value::OptionalValue;
use crate::{Context, Result};
use crate::error::Error;
use crate::message::Message;

/// Dynamically invokes an account message.
/// Static account client instances should be preferred wherever possible,
/// so that static dependency analysis can be performed.
pub unsafe fn dynamic_invoke<'a, 'b, M: Message<'b>>(context: &'a Context, account: AccountID, message: M)
                                                     -> crate::Result<<M::Response<'a> as OptionalValue<'a>>::Value> {
    // encode the message body
    let mem = context.memory_manager();
    let cdc = M::Codec::default();
    let msg_body = cdc.encode_value(&message, mem)
        // map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        .map_err(|_| ())?;

    // create the message packet and fill in call details
    let mut packet = create_packet(context, account, M::SELECTOR)
        // .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        .map_err(|_| ())?;
    let header = packet.header_mut();
    header.in_pointer1.set_slice(msg_body);

    // invoke the message
    let res = context.host_backend().invoke(&mut packet, mem)
        .map_err(|_| ());

    let out1 = header.out_pointer1.get(&packet);

    match res {
        Ok(_) => {
            let res = M::Response::<'a>::decode_value(&cdc, &out1, mem)
                // map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
                .map_err(|_| ())?;
            Ok(res)
        }
        Err(_) => {
            //TODO
            Err(())
        }
    }
}

/// Create a new message packet with the given account and message selector.
pub fn create_packet<'a>(context: &'a Context, account: AccountID, selector: u64) -> Result<MessagePacket<'a>> {
    unsafe {
        let packet = MessagePacket::allocate(context.memory_manager(), 0)
            // .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
            .map_err(|_| ())?;
        let header = packet.header_mut();
        header.context_info.caller = context.account_id();
        header.context_info.account = account;
        header.message_selector = selector;
        Ok(packet)
    }
}

/// Encodes the optional value to the out1 pointer of the message packet. Used for encoding the response of a message in macros.
pub fn encode_optional_to_out1<'b, 'a, V: OptionalValue<'a>>(cdc: &dyn Codec, value: &V::Value, writer_factory: &'b dyn Allocator, message_packet: &'b mut MessagePacket) -> core::result::Result<(), EncodeError>{
    if let Some(out1) = V::encode_value(cdc, value, &writer_factory as &dyn WriterFactory)? {
        unsafe { message_packet.header_mut().out_pointer1.set_slice(out1); }
    }
    Ok(())
}
