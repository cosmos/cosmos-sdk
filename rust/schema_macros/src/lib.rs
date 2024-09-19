//! Macros for generating code for the schema crate.

use proc_macro2::TokenStream as TokenStream2;
use manyhow::{bail, manyhow};
use quote::quote;
use syn::Data;

/// This derives a struct codec.
#[manyhow]
#[proc_macro_derive(StructCodec)]
pub fn derive_struct_codec(input: syn::DeriveInput) -> manyhow::Result<TokenStream2> {
    let struct_name = input.ident;
    let str = match &input.data {
        Data::Struct(str) => {
            str
        }
        _ => {
            bail!("StructCodec can only be derived for structs");
        }
    };
    let dummy_impls = str.fields.iter().map(|field| {
        let field_name = field.ident.as_ref().unwrap();
        let field_type = &field.ty;
        quote! {
                    <#field_type as ::interchain_schema::value::MaybeBorrowed>::dummy();
                }
    });
    Ok(quote! {
        unsafe impl StructCodec for #struct_name {
            fn dummy(&self) {
                #(#dummy_impls)*
            }
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
