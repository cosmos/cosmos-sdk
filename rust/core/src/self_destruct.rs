//! Self-destruct functionality for accounts.

use crate::context::Context;
use crate::response::Response;

/// Self-destructs the account.
///
/// SAFETY: This function is unsafe because it can be used to destroy the account and all its state.
pub unsafe fn self_destruct<'a, 'b>(ctx: &mut Context<'a>) -> Response<'b, ()> {
    unimplemented!()
}
