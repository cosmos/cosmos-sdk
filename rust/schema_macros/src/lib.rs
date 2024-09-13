//! Macros for generating code for the schema crate.
use proc_macro::TokenStream;

/// This derives a struct codec.
#[proc_macro_derive(StructCodec)]
pub fn derive_struct_codec(input: TokenStream) -> TokenStream {
    TokenStream::new()
}