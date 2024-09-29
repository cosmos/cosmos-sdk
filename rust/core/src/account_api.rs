//! Self-destruct functionality for accounts.

use ixc_core_macros::message_selector;
use crate::context::Context;
use crate::response::Response;

/// Self-destructs the account.
///
/// SAFETY: This function is unsafe because it can be used to destroy the account and all its state.
pub unsafe fn self_destruct<'a, 'b>(ctx: &mut Context<'a>) -> Response<'b, ()> {
    unimplemented!()
}

/// The selector for the account create message.
pub const CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.create");
