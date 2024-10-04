//! **WARNING: This is an API preview! Most code won't work or even type check properly!**
//! Macros for generating code for the schema crate.

use manyhow::{bail, manyhow};
use proc_macro2::{Ident, Span, TokenStream as TokenStream2};
use quote::quote;
use syn::{Attribute, Data, DataStruct, Lifetime};

/// This derives a struct codec.
#[manyhow]
#[proc_macro_derive(SchemaValue, attributes(sealed, schema, proto))]
pub fn derive_schema_value(input: syn::DeriveInput) -> manyhow::Result<TokenStream2> {
    match &input.data {
        Data::Struct(str) => {
            return derive_struct_schema(&input, str);
        }
        _ => bail!("only know how to derive SchemaValue for structs")
    }
}

fn derive_struct_schema(input: &syn::DeriveInput, str: &DataStruct) -> manyhow::Result<TokenStream2> {
    let struct_name = &input.ident;
    // extract struct lifetime
    let generics = &input.generics;
    if generics.lifetimes().count() > 1 {
        bail!("only one lifetime parameter is allowed")
    }
    let (impl_generics, ty_generics, where_clause) = generics.split_for_impl();
    let lifetime = if let Some(lifetime) = generics.lifetimes().next() {
        lifetime.lifetime.clone()
    } else {
        Lifetime::new("'a", Span::call_site())
    };
    let lifetime2 = if lifetime.ident == "b" {
        Lifetime::new("'c", Span::call_site())
    } else {
        Lifetime::new("'b", Span::call_site())
    };
    let ty_generics2 = if let Some(lifetime) = generics.lifetimes().next() {
        quote! { < #lifetime2 > }
    } else {
        quote! {}
    };

    let sealed = has_attribute(&input.attrs, "sealed");
    let non_exhaustive = has_attribute(&input.attrs, "non_exhaustive");
    if !sealed && !non_exhaustive {
        bail!("struct must have either a #[sealed] or #[non_exhaustive] attribute to indicate whether adding new fields is or is not a breaking change")
    }
    if sealed && non_exhaustive {
        bail!("struct cannot be both sealed and non_exhaustive")
    }

    let fields = str.fields.iter().map(|field| {
        let field_name = field.ident.as_ref().unwrap();
        let field_type = &field.ty;
        quote! {
               ::ixc_schema::types::to_field::<<#field_type as ::ixc_schema::SchemaValue< '_ >>::Type>().with_name(stringify!(#field_name)),
        }
    });
    let encode_matchers = str.fields.iter().enumerate().map(|(index, field)| {
        let field_name = field.ident.as_ref().unwrap();
        let field_type = &field.ty;
        quote! {
            #index => <#field_type as ::ixc_schema::SchemaValue<'_>>::encode(&self.#field_name, encoder),
        }
    });
    let decode_states = str.fields.iter().map(|field| {
        let field_type = &field.ty;
        quote! {
            <#field_type as ::ixc_schema::SchemaValue< #lifetime >>::DecodeState,
        }
    });
    let decode_matchers = str.fields.iter().enumerate().map(|(index, field)| {
        let field_type = &field.ty;
        let tuple_index = syn::Index::from(index);
        quote! {
            #index => <#field_type as ::ixc_schema::SchemaValue< #lifetime >>::visit_decode_state(&mut self.state.#tuple_index, decoder),
        }
    });
    let finishers = str.fields.iter().enumerate().map(|(index, field)| {
        let field_name = field.ident.as_ref().unwrap();
        let field_type = &field.ty;
        let tuple_index = syn::Index::from(index);
        quote! {
            let #field_name = <#field_type as ::ixc_schema::SchemaValue< #lifetime >>::finish_decode_state(state.#tuple_index, mem)?;
        }
    });
    let field_inits = str.fields.iter().enumerate().map(|(index, field)| {
        let field_name = field.ident.as_ref().unwrap();
        quote! {
            #field_name,
        }
    });
    Ok(quote! {
        unsafe impl #impl_generics ::ixc_schema::structs::StructSchema for #struct_name #ty_generics #where_clause {
            const STRUCT_TYPE: ::ixc_schema::structs::StructType<'static> = ::ixc_schema::structs::StructType {
                name: stringify!(#struct_name),
                fields: &[#(#fields)*],
                sealed: #sealed,
            };
        }

        unsafe impl #impl_generics ::ixc_schema::types::ReferenceableType for #struct_name #ty_generics #where_clause {
            const SCHEMA_TYPE: Option<::ixc_schema::schema::SchemaType<'static>> = Some(
                ::ixc_schema::schema::SchemaType::Struct(<Self as ::ixc_schema::structs::StructSchema>::STRUCT_TYPE)
            );
        }

        unsafe impl #impl_generics ::ixc_schema::structs::StructEncodeVisitor for #struct_name #ty_generics #where_clause {
            fn encode_field(&self, index: usize, encoder: &mut dyn ::ixc_schema::encoder::Encoder) -> ::core::result::Result<(), ::ixc_schema::encoder::EncodeError> {
                match index {
                    #(#encode_matchers)*
                    _ => Err(::ixc_schema::encoder::EncodeError::UnknownError),
                }
            }
        }

        impl < #lifetime > ::ixc_schema::SchemaValue < #lifetime > for #struct_name #ty_generics #where_clause {
            type Type = ::ixc_schema::types::StructT< #struct_name #ty_generics >;
            type DecodeState = (#(#decode_states)*);

            fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn ::ixc_schema::decoder::Decoder< #lifetime >) -> ::core::result::Result<(), ::ixc_schema::decoder::DecodeError> {
                struct Visitor< #lifetime2 , #lifetime : #lifetime2 > {
                    state: &#lifetime2 mut < #struct_name #ty_generics as ::ixc_schema::SchemaValue< #lifetime >>::DecodeState,
                }
                unsafe impl< #lifetime2, #lifetime : #lifetime2 > ::ixc_schema::structs::StructDecodeVisitor< #lifetime > for Visitor< #lifetime2, #lifetime > {
                    fn decode_field(&mut self, index: usize, decoder: &mut dyn ::ixc_schema::decoder::Decoder< #lifetime >) -> ::core::result::Result<(), ::ixc_schema::decoder::DecodeError> {
                        match index {
                            #(#decode_matchers)*
                            _ => Err(::ixc_schema::decoder::DecodeError::UnknownFieldNumber),
                        }
                    }
                }
                decoder.decode_struct(&mut Visitor { state }, &<Self as ::ixc_schema::structs::StructSchema>::STRUCT_TYPE)
            }

            fn finish_decode_state(state: Self::DecodeState, mem: &#lifetime ::ixc_schema::mem::MemoryManager) -> ::core::result::Result<Self, ::ixc_schema::decoder::DecodeError> {
                #(#finishers)*
                Ok( #struct_name {
                    #(#field_inits)*
                })
            }

            fn encode(&self, encoder: &mut dyn ::ixc_schema::encoder::Encoder) -> ::core::result::Result<(), ::ixc_schema::encoder::EncodeError> {
                encoder.encode_struct(self, &<Self as ::ixc_schema::structs::StructSchema>::STRUCT_TYPE)
            }
        }

        // impl < #lifetime > ::ixc_schema::SchemaValue < #lifetime > for &#lifetime #struct_name #ty_generics #where_clause {
        //     type Type = ::ixc_schema::types::StructT< #struct_name #ty_generics >;
        // }

        impl < #lifetime > ::ixc_schema::value::ListElementValue < #lifetime > for #struct_name #ty_generics #where_clause {}
        impl #impl_generics ::ixc_schema::state_object::ObjectFieldValue for #struct_name #ty_generics #where_clause {
            type In< #lifetime2 > = #struct_name #ty_generics2;
            type Out< #lifetime2 > = #struct_name #ty_generics2;
        }
    }.into())
}

fn has_attribute<I>(attrs: &Vec<Attribute>, ident: &I) -> bool
where
    I: ?Sized,
    Ident: PartialEq<I>,
{
    for attr in attrs {
        if attr.path().is_ident(ident) {
            return true;
        }
    }
    false
}