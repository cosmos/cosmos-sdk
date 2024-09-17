//! Macros for generating code for the schema crate.
use proc_macro::TokenStream;

/// This derives a struct codec.
#[proc_macro_derive(StructCodec)]
pub fn derive_struct_codec(input: TokenStream) -> TokenStream {
    TokenStream::new()
}

/// This derives an enum codec.
#[proc_macro_derive(EnumCodec)]
pub fn derive_enum_codec(input: TokenStream) -> TokenStream {
    TokenStream::new()
}

/// This derives a oneof codec.
#[proc_macro_derive(OneOfCodec)]
pub fn derive_oneof_codec(input: TokenStream) -> TokenStream {
    TokenStream::new()
}
