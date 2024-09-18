//! This is a macro utility crate for interchain_core.

use proc_macro::{TokenStream};
use std::default::Default;
use proc_macro2::TokenStream as TokenStream2;
use manyhow::{bail, error_message, manyhow};
use quote::{format_ident, quote, ToTokens};
use syn::{parse2, parse_macro_input, token, Attribute, File, Item, ItemImpl, ItemMod, Type};
use std::borrow::Borrow;
use syn::token::Impl;

#[derive(deluxe::ParseMetaItem)]
struct HandlerArgs(syn::Ident);

/// This derives an account handler.
#[manyhow]
#[proc_macro_attribute]
pub fn account_handler(attr: TokenStream2, mut item: ItemMod) -> manyhow::Result<TokenStream2> {
    let HandlerArgs(handler) = deluxe::parse2(attr)?;
    let items = &mut item.content.as_mut().unwrap().1;
    push_item(items, quote! {
        impl ::interchain_core::handler::Handler for #handler {
            type Init = ();
        }
    })?;
    let client_ident = format_ident!("{}Client", handler);
    push_item(items, quote! {
        pub struct #client_ident(::interchain_message_api::Address);
    })?;
    push_item(items, quote! {
        impl ::interchain_core::handler::AccountClient for #client_ident {
            fn address(&self) -> &::interchain_message_api::Address {
                &self.0
            }
        }
    })?;
    let client_factory_ident = format_ident!("{}Factory", handler);
    push_item(items, quote! {
        pub struct #client_factory_ident;
    })?;
    let mut publish_targets = vec![];
    for item in items.iter_mut() {
        collect_publish_targets(&handler, item, &mut publish_targets)?;
    }

    let mut client_fn_impls = vec![];
    for publish_target in publish_targets.iter() {
        // if publish_target.on_create.is_some() {
        //     continue;
        // }
        let signature = &publish_target.signature;
        let fn_name = &signature.ident;
        client_fn_impls.push(quote! {
            pub fn #fn_name(&self, ctx: &::interchain_core::Context) -> ::interchain_core::Response<()> {
                todo!()
            }
        });
    }

    push_item(items, quote! {
        impl #client_ident {
            #(#client_fn_impls)*
        }
    })?;

    let expanded = quote! {
        #item
    };
    Ok(expanded)
}

fn push_item(item: &mut Vec<Item>, expanded: TokenStream2) -> manyhow::Result<()> {
    item.push(parse2::<Item>(expanded)?);
    Ok(())
}

fn collect_publish_targets(self_name: &syn::Ident, item: &mut Item, targets: &mut Vec<PublishFn>) -> manyhow::Result<()> {
    match item {
        Item::Impl(imp) => {
            match imp.self_ty.borrow() {
                Type::Path(self_path) => {
                    let ident = match self_path.path.get_ident() {
                        None => return Ok(()),
                        Some(ident) => ident,
                    };
                    if ident != self_name {
                        return Ok(());
                    }

                    // TODO check for trait implementation

                    let publish_all = deluxe::extract_attributes::<ItemImpl, Publish>(imp).map_or(None, |x| Some(x));
                    for item in &mut imp.items {
                        match item {
                            syn::ImplItem::Fn(impl_fn) => {
                                let on_create = deluxe::extract_attributes::<syn::ImplItemFn, OnCreate>(impl_fn).map_or(None, |x| Some(x));
                                let publish = deluxe::extract_attributes::<syn::ImplItemFn, Publish>(impl_fn).map_or(None, |x| Some(x));
                                let publish = publish_all.clone().or(publish);
                                if publish.is_some() || on_create.is_some() { // TODO check visibility
                                    targets.push(PublishFn {
                                        signature: impl_fn.sig.clone(),
                                        on_create,
                                        publish: publish.unwrap(),
                                    });
                                }
                            }
                            _ => {}
                        }
                    }
                }
                _ => {}
            }
        }
        _ => {}
    }
    Ok(())
}

#[derive(deluxe::ExtractAttributes, Clone)]
#[deluxe(attributes(publish))]
struct Publish {}

#[derive(deluxe::ExtractAttributes)]
#[deluxe(attributes(on_create))]
struct OnCreate {}

struct PublishFn {
    signature: syn::Signature,
    on_create: Option<OnCreate>,
    publish: Publish,
}

/// This derives an module handler.
#[proc_macro_attribute]
pub fn module_handler(_attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}

/// This publishes a trait or struct impl block or a single fn within an impl block.
#[manyhow]
#[proc_macro_attribute]
pub fn publish(_attr: TokenStream2, item: TokenStream2) -> manyhow::Result<TokenStream2> {
    bail!("the #[publish] attribute is being used in the wrong context, possibly #[module_handler] or #[account_handler] has not been applied to the enclosing module")
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

/// This attribute macro should be attached to the fn which is called when an account is created.
#[manyhow]
#[proc_macro_attribute]
pub fn on_create(_attr: TokenStream2, item: TokenStream2) -> manyhow::Result<TokenStream2> {
    bail!("the #[on_create] attribute is being used in the wrong context, possibly #[module_handler] or #[account_handler] has not been applied to the enclosing module")
}

/// Derive the `Resources` trait for a struct.
#[proc_macro_derive(Resources, attributes(schema, client))]
pub fn derive_resources(input: TokenStream) -> TokenStream {
    let input = parse_macro_input!(input as syn::DeriveInput);
    let name = input.ident;
    let expanded = quote! {
        unsafe impl ::interchain_core::resource::Resources for #name {
        }
    };
    expanded.into()
}

/// This attribute bundles account and module handlers into a package root which can be
/// loaded into an application.
#[proc_macro]
pub fn package_root(item: TokenStream) -> TokenStream {
    // let item = parse_macro_input!(item as File);
    // let expanded = quote! {
    //     #item
    // };
    // expanded.into()
    TokenStream::default()
}
