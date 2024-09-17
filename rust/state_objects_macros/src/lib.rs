//! Macros for state_objects.
use proc_macro::TokenStream;
use quote::quote;
use syn::{parse_macro_input, Field};

struct MapArgs {
    key: Vec<String>,
    value: Vec<String>,
    prefix: u8,
}

struct SetArgs {
    key: Vec<String>,
    prefix: u8,
}

struct ItemArgs {
    value: Vec<String>,
    prefix: u8,
}

struct IndexArgs {
    on: String, // on(some_map(key1, key2))
    prefix: u8,
}

/// Derive the `Resources` trait for a struct.
#[proc_macro_derive(Resources, attributes(map, set, item, index, unique_index, seq))]
pub fn derive_resources(input: TokenStream) -> TokenStream {
    let input = parse_macro_input!(input as syn::DeriveInput);
    let name = input.ident;
    let expanded = quote! {
        impl ::interchain_core::resource::Resources for #name {
        }
    };
    expanded.into()
}