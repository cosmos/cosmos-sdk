//! Self-destruct functionality for accounts.

use ixc_core_macros::message_selector;
use ixc_message_api::AccountID;
use ixc_message_api::handler::{HandlerErrorCode, Allocator};
use ixc_schema::codec::Codec;
use ixc_schema::value::OptionalValue;
use crate::context::Context;
use crate::error::Error;
use crate::handler::Handler;

pub fn create_account<'a, H: Handler>(ctx: &mut Context, init: &<H::Init<'a> as OptionalValue<'a>>::Value) -> crate::Result<H::Client> {
    let packet = ctx.memory_manager().allocate_packet(0)
        .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;

    <H::Init<'_> as OptionalValue<'_>>::encode_value::<H::InitCodec>(init, packet, ctx.memory_manager() as &dyn Allocator).
        map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;

    let header = packet.header_mut();
    header.sender_account = ctx.account_id();
    header.account = HYPERVISOR_ACCOUNT;
    header.message_selector = CREATE_SELECTOR;
    unsafe {
        ctx.host_backend().invoke(packet, ctx.memory_manager())
            .map_err(|_| todo!())?
    }
    // TODO decode the account ID from the response
    Ok(H::new_client(AccountID::new(0)))
}

/// Self-destructs the account.
///
/// SAFETY: This function is unsafe because it can be used to destroy the account and all its state.
pub unsafe fn self_destruct(ctx: &mut Context) -> crate::Result<()> {
    unimplemented!()
}

const HYPERVISOR_ACCOUNT: AccountID = AccountID::new(1);

const CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.create");
