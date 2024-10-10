//! The core account API.

use crate::context::Context;
use crate::handler::{Service, Handler, InitMessage, Client};
use crate::low_level::create_packet;
use ixc_core_macros::message_selector;
use ixc_message_api::AccountID;
use ixc_schema::codec::Codec;
use crate::result::ClientResult;

/// Creates a new account for the specified handler.
pub fn create_account<'a, I: InitMessage<'a>>(ctx: &mut Context, init: I) -> ClientResult<<<I as InitMessage<'a>>::Handler as Service>::Client> {
    let cdc = I::Codec::default();
    let init_bz = cdc.encode_value(&init, ctx.memory_manager())?;

    let account_id = do_create_account(ctx, I::Handler::NAME, &init_bz)?;
    Ok(<<I::Handler as Service>::Client as Client>::new(account_id))
}

/// Creates a new account for the named handler with opaque initialization data.
pub fn create_account_raw<'a>(ctx: &mut Context, name: &str, init: &[u8]) -> ClientResult<AccountID> {
    do_create_account(ctx, name, init)
}

/// Creates a new account for the named handler with opaque initialization data.
fn do_create_account<'a>(ctx: &Context, name: &str, init: &[u8]) -> ClientResult<AccountID> {
    let mut packet = create_packet(ctx, ROOT_ACCOUNT, CREATE_SELECTOR)?;

    unsafe {
        packet.header_mut().in_pointer1.set_slice(name.as_bytes());
        packet.header_mut().in_pointer2.set_slice(init);

        ctx.host_backend().invoke(&mut packet, ctx.memory_manager())?;

        let new_account_id = packet.header().in_pointer1.get_u64();

        Ok(AccountID::new(new_account_id))
    }
}

/// Self-destructs the account.
///
/// SAFETY: This function is unsafe because it can be used to destroy the account and all its state.
pub unsafe fn self_destruct(ctx: &mut Context) -> ClientResult<()> {
    let mut packet = create_packet(ctx, ROOT_ACCOUNT, SELF_DESTRUCT_SELECTOR)?;
    unsafe {
        ctx.host_backend().invoke(&mut packet, ctx.memory_manager())?;
        Ok(())
    }
}

const CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.create");

const SELF_DESTRUCT_SELECTOR: u64 = message_selector!("ixc.account.v1.self_destruct");

/// The ID of the root account which creates and manages accounts.
pub const ROOT_ACCOUNT: AccountID = AccountID::new(1);

/// The message selector for the on_create message.
pub const ON_CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.on_create");

// #[ixc_schema_macros::handler_api]
/// The API for converting between native addresses and account IDs.
/// Native addresses have both a byte representation and a string representation.
/// The mapping between addresses and account IDs is assumed to be stateful.
pub trait AddressAPI {
    /// Convert an account ID to a byte representation.
    fn to_bytes<'a>(&self, ctx: &'a Context, account_id: AccountID) -> crate::Result<&'a [u8]>;
    /// Convert a byte representation to an account ID.
    fn from_bytes<'a>(&self, ctx: &'a Context, address_bytes: &[u8]) -> crate::Result<AccountID>;
    /// Convert an account ID to a string representation.
    fn to_string<'a>(&self, ctx: &'a Context, account_id: AccountID) -> crate::Result<&'a str>;
    /// Convert a string representation to an account ID.
    fn from_string<'a>(&self, ctx: &'a Context, address_string: &str) -> crate::Result<AccountID>;
}