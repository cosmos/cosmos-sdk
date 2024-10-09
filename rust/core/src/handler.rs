//! Handler traits for account and module handlers.
use crate::resource::Resources;
use crate::routes::Router;
use ixc_message_api::handler::RawHandler;
use ixc_message_api::AccountID;
use ixc_schema::codec::Codec;
use ixc_schema::structs::StructSchema;
use ixc_schema::SchemaValue;

/// Handler trait for account and module handlers.
pub trait Handler: RawHandler + Router + Resources + ClientFactory {
    /// The name of the handler.
    const NAME: &'static str;
    /// The parameter used for initializing the handler.
    type Init<'a>: InitMessage<'a>;
}

/// A message which initializes a new account for a handler.
pub trait InitMessage<'a>: SchemaValue<'a> + StructSchema
// TODO required a sealed struct
{
    /// The handle which the account will be created with.
    type Handler: Handler;
    /// The codec used for initializing the handler.
    type Codec: Codec + Default;
}

/// Account API trait.
pub trait HandlerAPI: Router {
    /// Account client factory type.
    type ClientFactory: ClientFactory;
}

/// Account factory trait.
pub trait ClientFactory {
    /// Account client type.
    type Client: Client;

    /// Create a new account client with the given address.
    fn new_client(account_id: AccountID) -> Self::Client;
}

/// Account client trait.
pub trait Client {
    /// Get the address of the account.
    fn account_id(&self) -> AccountID;
}

/// The client of a handler.
pub trait HandlerClient: Client {
    /// The handler type.
    type Handler: Handler;
}
