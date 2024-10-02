use allocator_api2::alloc::Allocator;
use ixc_message_api::AccountID;
use ixc_message_api::handler::{HandlerErrorCode};
use ixc_message_api::packet::MessagePacket;
use ixc_schema::codec::Codec;
use ixc_schema::value::OptionalValue;
use crate::{Context, Result};
use crate::error::Error;
use crate::message::Message;

/// Dynamically invokes an account message.
/// Static account client instances should be preferred wherever possible,
/// so that static dependency analysis can be performed.
pub unsafe fn dynamic_invoke<'a, 'b, M: Message<'b>>(context: &'a Context, account: AccountID, message: M)
                                                 -> crate::Result<<M::Response<'a> as OptionalValue<'a>>::Value, M::Error> {
    // encode the message body
    let mem = context.memory_manager();
    let msg_body = M::Codec::encode_value(&message, mem as &dyn Allocator).
        map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;

    // create the message packet and fill in call details
    let packet = create_packet(context, account, M::SELECTOR)
        .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
    let header = packet.header_mut();
    header.in_pointer1.set_slice(msg_body);

    // invoke the message
    let res = context.host_backend().invoke(packet, mem)
        .map_err(|_| todo!());

    match res {
        Ok(_) => {
            let res = M::Response::<'a>::decode_value::<M::Codec>(packet, mem).
                map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
            Ok(res)
        }
        Err(_) => {
            todo!()
        }
    }
}

pub fn create_packet<'a>(context: &'a Context, account: AccountID, selector: u64) -> Result<&'a mut MessagePacket> {
    let packet = context.memory_manager().allocate_packet(0)
        .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
    let header = packet.header_mut();
    header.sender_account = context.account_id();
    header.account = account;
    header.message_selector = selector;
    Ok(packet)
}