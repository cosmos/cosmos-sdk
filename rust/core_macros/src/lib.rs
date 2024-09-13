//! This is a macro utility crate for interchain_core.

use proc_macro::{TokenStream};
use quote::quote;
use syn::{parse_macro_input, File, Item};

/// This derives an account handler.
#[proc_macro_attribute]
pub fn account_handler(_attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}

/// This derives an module handler.
#[proc_macro_attribute]
pub fn module_handler(_attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}

/// This publishes a trait or struct impl block or a single fn within an impl block.
#[proc_macro_attribute]
pub fn publish(_attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}

/// This attribute macro should be attached to a trait that implements a account API.
#[proc_macro_attribute]
pub fn account_api(_attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}

/// This attribute macro should be attached to a trait that implements a module API.
#[proc_macro_attribute]
pub fn module_api(_attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}

/// This attribute bundles account and module handlers into a package root which can be
/// loaded into an application.
#[proc_macro_attribute]
pub fn package(_attr: TokenStream, item: TokenStream) -> TokenStream {
    let item = parse_macro_input!(item as File);
    let expanded = quote! {
        #item
    };
    expanded.into()
}

/// This attribute macro should be attached to a function that is called when the account is created.
#[proc_macro_attribute]
pub fn on_create(_attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}

/// This attribute macro should be attached to a function that is called when the account is migrated.
#[proc_macro_attribute]
pub fn on_migrate(_attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}
