//! **WARNING: This is an API preview! Most code won't work or even type check properly!**
//! Macros for generating code for the schema crate.

use proc_macro2::TokenStream as TokenStream2;
use manyhow::{bail, manyhow};
use quote::quote;
use syn::Data;

/// This derives a struct codec.
#[manyhow]
#[proc_macro_derive(SchemaValue)]
pub fn derive_schema_value(input: syn::DeriveInput) -> manyhow::Result<TokenStream2> {
    let struct_name = input.ident;
    let str = match &input.data {
        Data::Struct(str) => {
            str
        }
        _ => {
            bail!("only know how to derive SchemaValue for structs, currently");
        }
    };
    let dummy_impls = str.fields.iter().map(|field| {
        let field_name = field.ident.as_ref().unwrap();
        let field_type = &field.ty;
        quote! {

        }
    });
    Ok(quote! {
        unsafe impl <'a> SchemaValue<'a> for #struct_name {
            type Type = ::ixc_schema::types::StructT(#struct_name);
        }
    }.into())
}

/// This derives an enum codec.
#[manyhow]
#[proc_macro_derive(EnumCodec)]
pub fn derive_enum_codec(input: TokenStream2) -> manyhow::Result<TokenStream2> {
    Ok(TokenStream2::new())
}

/// This derives a oneof codec.
#[manyhow]
#[proc_macro_derive(OneOfCodec)]
pub fn derive_oneof_codec(input: TokenStream2) -> manyhow::Result<TokenStream2> {
    Ok(TokenStream2::new())
}
