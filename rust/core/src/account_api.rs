//! Self-destruct functionality for accounts.

use ixc_core_macros::message_selector;
use ixc_message_api::AccountID;
use ixc_message_api::code::SystemErrorCode;
use ixc_message_api::handler::{HandlerErrorCode, Allocator};
use ixc_schema::codec::Codec;
use ixc_schema::value::OptionalValue;
use crate::context::Context;
use crate::error::Error;
use crate::error::Error::SystemError;
use crate::handler::Handler;
use crate::low_level::create_packet;

/// Creates a new account for the specified handler.
pub fn create_account<'a, H: Handler>(ctx: &mut Context, init: &<H::Init<'a> as OptionalValue<'a>>::Value) -> crate::Result<H::Client> {
    let packet = create_packet(ctx, HYPERVISOR_ACCOUNT, CREATE_SELECTOR)?;

    <H::Init<'_> as OptionalValue<'_>>::encode_value::<H::InitCodec>(init, packet, ctx.memory_manager() as &dyn Allocator).
        map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;

    unsafe {
        ctx.host_backend().invoke(packet, ctx.memory_manager())
            .map_err(|_| Error::SystemError(SystemErrorCode::UnknownHandlerError))?;

        let new_account_id = packet.header().in_pointer1.get_u64();

        Ok(H::new_client(AccountID::new(new_account_id)))
    }
}

/// Self-destructs the account.
///
/// SAFETY: This function is unsafe because it can be used to destroy the account and all its state.
pub unsafe fn self_destruct(ctx: &mut Context) -> crate::Result<()> {
    unimplemented!()
}

const HYPERVISOR_ACCOUNT: AccountID = AccountID::new(1);

const CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.create");
