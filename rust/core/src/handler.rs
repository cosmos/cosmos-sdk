//! Handler traits for account and module handlers.
use crate::resource::Resources;
use crate::routing::Router;
use ixc_message_api::handler::RawHandler;
use ixc_message_api::AccountID;
use ixc_schema::codec::Codec;
use ixc_schema::structs::StructSchema;
use ixc_schema::SchemaValue;

/// Handler trait for account and module handlers.
pub trait Handler: RawHandler + Router + Resources + Service {
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

/// Something that implements a service and thus has an associated client type.
pub trait Service {
    /// The client type associated with the service.
    type Client: Client;

    /// Create a new client for the service with the given account ID.
    fn new_client(account_id: AccountID) -> Self::Client {
        Self::Client::new(account_id)
    }
}

/// The trait that clients of a service must implement.
pub trait Client {
    /// Create a new client with the given account ID.
    fn new(account_id: AccountID) -> Self;

    /// Get the address of the account this client is associated with.
    fn account_id(&self) -> AccountID;
}

/// The client of a handler.
pub trait HandlerClient: Client {
    /// The handler type.
    type Handler: Handler;
}
