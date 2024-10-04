//! The core account API.

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
    let mut packet = create_packet(ctx, ROOT_ACCOUNT, CREATE_SELECTOR)?;

    let cdc = H::InitCodec::default();
    let init_bz = <H::Init<'_> as OptionalValue<'_>>::encode_value(&cdc, init, ctx.memory_manager()).
        // map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        map_err(|_| ())?;


    unsafe {
        packet.header_mut().in_pointer1.set_slice(H::NAME.as_bytes());
        if let Some(init_bz) = init_bz {
            packet.header_mut().in_pointer2.set_slice(init_bz);
        }

        ctx.host_backend().invoke(&mut packet, ctx.memory_manager())
            // .map_err(|_| Error::SystemError(SystemErrorCode::UnknownHandlerError))?;
            .map_err(|_| ())?;

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

const CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.create");

/// The ID of the root account which creates and manages accounts.
pub const ROOT_ACCOUNT: AccountID = AccountID::new(1);

/// The message selector for the on_create message.
pub const ON_CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.on_create");
