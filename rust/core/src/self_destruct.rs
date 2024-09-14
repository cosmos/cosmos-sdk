//! Self-destruct functionality for accounts.
use crate::{Context, Response};

/// Self-destructs the account.
///
/// SAFETY: This function is unsafe because it can be used to destroy the account and all its state.
pub unsafe fn self_destruct(ctx: &mut Context) -> Response<()> {
    unimplemented!()
}
