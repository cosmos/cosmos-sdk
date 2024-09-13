//! This is a macro utility crate for interchain_core.

use proc_macro::TokenStream;

/// This derives an account handler.
#[proc_macro]
pub fn account_handler(item: TokenStream) -> TokenStream {
    item
}

/// This derives an module handler.
#[proc_macro]
pub fn module_handler(item: TokenStream) -> TokenStream {
    item
}

/// This publishes a trait or struct impl block or a single fn within an impl block.
#[proc_macro]
pub fn publish(item: TokenStream) -> TokenStream {
    item
}

/// This attribute macro should be attached to a trait that implements a account API.
#[proc_macro]
pub fn account_api(item: TokenStream) -> TokenStream {
    item
}

/// This attribute macro should be attached to a trait that implements a module API.
#[proc_macro]
pub fn module_api(item: TokenStream) -> TokenStream {
    item
}
