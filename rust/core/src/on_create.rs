use std::task::Context;
use interchain_schema::StructCodec;
use crate::Response;

/// The trait that account handlers define to initialize themselves.
pub trait OnCreate {
    /// The message type that is used to initialize the account.
    type InitMessage;

    /// Initializes the account with the given message.
    fn on_create(&self, ctx: &mut Context, init: &Self::InitMessage) -> Response<()>;
}