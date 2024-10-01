//! **WARNING: This is an API preview! Most code won't work or even type check properly!**
//! Macros for generating code for the schema crate.

use proc_macro2::{Span, TokenStream as TokenStream2};
use manyhow::{bail, manyhow};
use quote::quote;
use syn::{Data, Lifetime, Member};

/// This derives a struct codec.
#[manyhow]
#[proc_macro_derive(SchemaValue, attributes(sealed))]
pub fn derive_schema_value(input: syn::DeriveInput) -> manyhow::Result<TokenStream2> {
    let struct_name = input.ident;
    let str = match &input.data {
        Data::Struct(str) => {
            str
        }
        _ => bail!("only know how to derive SchemaValue for structs, currently")
    };

    // extract struct lifetime
    let mut generics = input.generics;
    if generics.lifetimes().count() > 1 {
        bail!("only one lifetime parameter is allowed")
    }
    let lifetime = if let Some(lifetime) = generics.lifetimes().next() {
        lifetime.lifetime.clone()
    } else {
        Lifetime::new("'a", Span::call_site())
    };
    let (impl_generics, ty_generics, where_clause) = generics.split_for_impl();

    // TODO: extract either non_exhaustive or sealed attribute
    let fields = str.fields.iter().map(|field| {
        let field_name = field.ident.as_ref().unwrap();
        let field_type = &field.ty;
        quote! {
                to_field::<<#field_type as SchemaValue<'a>>::Type>().with_name(stringify!(#field_name)),
        }
    });
    let encode_matchers = str.fields.iter().enumerate().map(|(index, field)| {
        let field_name = field.ident.as_ref().unwrap();
        let field_type = &field.ty;
        quote! {
            #index => <#field_type as SchemaValue<'_>>::encode(&self.#field_name, encoder),
        }
    });
    let decode_states = str.fields.iter().map(|field| {
        let field_type = &field.ty;
        quote! {
            <#field_type as SchemaValue<'a>>::DecodeState,
        }
    });
    let decode_matchers = str.fields.iter().enumerate().map(|(index, field)| {
        let field_type = &field.ty;
        let tuple_index = syn::Index::from(index);
        quote! {
            #index => <#field_type as SchemaValue< #lifetime >>::visit_decode_state(&mut self.state.#tuple_index, decoder),
        }
    });
    let finishers = str.fields.iter().enumerate().map(|(index, field)| {
        let field_name = field.ident.as_ref().unwrap();
        let field_type = &field.ty;
        let tuple_index = syn::Index::from(index);
        quote! {
            let #field_name = <#field_type as SchemaValue<'a>>::finish_decode_state(state.#tuple_index, mem)?;
        }
    });
    let field_inits = str.fields.iter().enumerate().map(|(index, field)| {
        let field_name = field.ident.as_ref().unwrap();
        quote! {
            #field_name,
        }
    });
    Ok(quote! {
        unsafe impl #impl_generics StructSchema for #struct_name #ty_generics #where_clause {
            const STRUCT_TYPE: StructType<'static> = StructType {
                name: stringify!(#struct_name),
                fields: &[#(#fields)*],
                sealed: false,
            };
        }

        unsafe impl #impl_generics ReferenceableType for #struct_name #ty_generics #where_clause {
            const SCHEMA_TYPE: Option<SchemaType<'static>> = Some(
                SchemaType::Struct(Self::STRUCT_TYPE)
            );
        }

        unsafe impl #impl_generics StructEncodeVisitor for #struct_name #ty_generics #where_clause {
            fn encode_field(&self, index: usize, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
                match index {
                    #(#encode_matchers)*
                    _ => Err(EncodeError::UnknownError),
                }
            }
        }

        impl < #lifetime > SchemaValue < #lifetime > for #struct_name #ty_generics #where_clause {
            type Type = StructT< #struct_name #ty_generics >;
            type DecodeState = (#(#decode_states)*);

            fn visit_decode_state(state: &mut Self::DecodeState, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
                struct Visitor<'b, #lifetime : 'b> {
                    state: &'b mut < #struct_name #ty_generics as SchemaValue<'a>>::DecodeState,
                }
                unsafe impl<'b, 'a: 'b> StructDecodeVisitor<'a> for Visitor<'b, 'a> {
                    fn decode_field(&mut self, index: usize, decoder: &mut dyn Decoder<'a>) -> Result<(), DecodeError> {
                        match index {
                            #(#decode_matchers)*
                            _ => Err(DecodeError::UnknownFieldNumber),
                        }
                    }
                }
                decoder.decode_struct(&mut Visitor { state }, &Self::STRUCT_TYPE)
            }

            fn finish_decode_state(state: Self::DecodeState, mem: &'a MemoryManager) -> Result<Self, DecodeError> {
                #(#finishers)*
                Ok( #struct_name {
                    #(#field_inits)*
                })
            }

            fn encode(&self, encoder: &mut dyn Encoder) -> Result<(), EncodeError> {
                encoder.encode_struct(self, &Self::STRUCT_TYPE)
            }
        }

        impl < #lifetime > ListElementValue < #lifetime > for #struct_name #ty_generics #where_clause {}
        // TODO: need to change lifetime 'b
        // impl < #lifetime > ObjectFieldValue < #lifetime > for #struct_name #ty_generics #where_clause {
        //     type In<'b> = Coin<'b>;
        //     type Out<'b> = Coin<'b>;
        // }
    }.into())
}
