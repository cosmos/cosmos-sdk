//! **WARNING: This is an API preview! Most code won't work or even type check properly!**
//! This is a macro utility crate for ixc_core.

use proc_macro::{TokenStream};
use std::default::Default;
use proc_macro2::{Ident, TokenStream as TokenStream2};
use manyhow::{bail, ensure, manyhow};
use quote::{format_ident, quote, ToTokens};
use syn::{parse2, parse_macro_input, parse_quote, Data, DeriveInput, Item, ItemMod, LitStr, ReturnType, Type};
use std::borrow::Borrow;
use blake2::{Blake2b512, Digest};
use deluxe::ExtractAttributes;
use heck::{ToUpperCamelCase};

#[derive(deluxe::ParseMetaItem)]
struct HandlerArgs(syn::Ident);

/// This derives an account handler.
#[manyhow]
#[proc_macro_attribute]
pub fn handler(attr: TokenStream2, mut item: ItemMod) -> manyhow::Result<TokenStream2> {
    let HandlerArgs(handler) = deluxe::parse2(attr)?;
    let items = &mut item.content.as_mut().unwrap().1;

    let mut publish_targets = vec![];
    for item in items.iter_mut() {
        collect_publish_targets(&handler, item, &mut publish_targets)?;
    }

    let client_ident = format_ident!("{}Client", handler);
    push_item(items, quote! {
        pub struct #client_ident(::ixc_message_api::AccountID);
    })?;
    push_item(items, quote! {
        impl ::ixc_core::handler::Client for #client_ident {
            fn account_id(&self) -> ::ixc_message_api::AccountID {
                self.0
            }
        }
    })?;
    push_item(items, quote! {
        impl ::ixc_core::handler::ClientFactory for #handler {
            type Client = #client_ident;

            fn new_client(account_id: ::ixc_message_api::AccountID) -> Self::Client {
                #client_ident(account_id)
            }
        }
    })?;
    push_item(items, quote! {
        impl ::ixc_message_api::handler::RawHandler for #handler {
            fn handle(&self, message_packet: &mut ::ixc_message_api::packet::MessagePacket, callbacks: &dyn ixc_message_api::handler::HostBackend, allocator: &dyn ::ixc_message_api::handler::Allocator) -> ::core::result::Result<(), ::ixc_message_api::handler::HandlerError> {
                ::ixc_core::routes::exec_route(self, message_packet, callbacks, allocator)
            }
        }
    })?;

    let mut client_fn_impls = vec![];
    let mut routes = vec![];
    let mut on_create_msg = quote! { () };
    for publish_target in publish_targets.iter() {
        let fn_name = &publish_target.signature.ident;
        let ident_camel = fn_name.to_string().to_upper_camel_case();
        let msg_struct_name = format_ident!("{}{}Msg", handler, ident_camel);
        let signature = publish_target.signature.clone();
        let mut msg_fields = vec![];
        let mut msg_fields_access = vec![];
        let mut msg_fields_init = vec![];
        let mut have_lifetimes = false;
        let mut context_name: Option<Ident> = None;
        for field in &publish_target.signature.inputs {
            match field {
                syn::FnArg::Typed(pat_type) => {
                    match pat_type.pat.as_ref() {
                        syn::Pat::Ident(ident) => {
                            let mut ty = pat_type.ty.clone();
                            match ty.as_mut() {
                                Type::Reference(tyref) => {
                                    if tyref.elem == parse_quote!(Context) {
                                        context_name = Some(ident.ident.clone());
                                        continue;
                                    }
                                    have_lifetimes = true;
                                    assert!(tyref.lifetime.is_none(), "support for named lifetimes is unimplemented");
                                    tyref.lifetime = Some(parse_quote!('a));
                                    assert!(false, "no support for borrowed data yet!")
                                }
                                _ => {}
                            }
                            msg_fields.push(quote! {
                                pub #ident: #ty,
                            });
                            msg_fields_access.push(quote! {
                                msg.#ident,
                            });
                            msg_fields_init.push(quote! {
                                #ident,
                            });
                        }
                        _ => bail!("expected identifier"),
                    };
                }
                _ => {}
            }
        }
        push_item(items, quote! {
            #[derive(::ixc_schema_macros::SchemaValue)]
            #[sealed]
            pub struct #msg_struct_name {
                #(#msg_fields)*
            }
        })?;
        let selector = message_selector_from_str(msg_struct_name.to_string().as_str());
        let return_type = match &signature.output {
            ReturnType::Type(_, ty) => ty,
            ReturnType::Default => {
                bail!("expected return type")
            }
        };
        if publish_target.on_create.is_none() {
            push_item(items, quote! {
                impl <'a> ::ixc_core::message::Message<'a> for #msg_struct_name {
                    const SELECTOR: ::ixc_message_api::header::MessageSelector = #selector;
                    type Response<'b> = <#return_type as ::ixc_core::message::ExtractResponseTypes>::Response;
                    type Error = ();
                    type Codec = ::ixc_schema::binary::NativeBinaryCodec;
                }
            })?;
            ensure!(context_name.is_some(), "no context parameter found");
            let context_name = context_name.unwrap();
            client_fn_impls.push(quote! {
                pub #signature {
                    let msg = #msg_struct_name {
                        #(#msg_fields_init)*
                    };
                    unsafe { ::ixc_core::low_level::dynamic_invoke(#context_name, self.0, msg) }
                }
            });
            routes.push(quote! {
                    (< #msg_struct_name as ::ixc_core::message::Message >::SELECTOR, |h: &#handler, packet, cb, a| {
                        unsafe {
                            let cdc = < #msg_struct_name as ::ixc_core::message::Message >::Codec::default();
                            let in1 = packet.header().in_pointer1.get(packet);
                            let mut ctx = ::ixc_core::Context::new(packet.header().context_info, cb);
                            let msg = ::ixc_schema::codec::decode_value::< #msg_struct_name >(&cdc, in1, ctx.memory_manager()).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))?;
                            let res = h.#fn_name(&mut ctx, #(#msg_fields_access)*).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))?;
                            ::ixc_core::low_level::encode_optional_to_out1::< < #msg_struct_name as ::ixc_core::message::Message<'_> >::Response<'_> >(&cdc, &res, a, packet).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))
                        }
                    }),
            });
        } else {
            on_create_msg = quote! { #msg_struct_name };
            routes.push(quote! {
                    (::ixc_core::account_api::ON_CREATE_SELECTOR, |h: &#handler, packet, cb, a| {
                        unsafe {
                            let cdc = < #handler as ::ixc_core::handler::Handler >::InitCodec::default();
                            let in1 = packet.header().in_pointer1.get(packet);
                            let mut ctx = ::ixc_core::Context::new(packet.header().context_info, cb);
                            let msg = ::ixc_schema::codec::decode_value::< #msg_struct_name >(&cdc, in1, ctx.memory_manager()).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))?;
                            h.#fn_name(&mut ctx, #(#msg_fields_access)*).map_err(|e| ::ixc_message_api::handler::HandlerError::Custom(0))
                        }
                    }),
            });
        }
    }

    push_item(items, quote! {
        impl ::ixc_core::handler::Handler for #handler {
            const NAME: &'static str = stringify!(#handler);
            type Init<'a> = #on_create_msg;
            type InitCodec = ::ixc_schema::binary::NativeBinaryCodec;
        }
    })?;

    push_item(items, quote! {
        impl #client_ident {
            #(#client_fn_impls)*
        }
    })?;

    push_item(items, quote! {
        unsafe impl ::ixc_core::routes::Router for #handler {
            const SORTED_ROUTES: &'static [::ixc_core::routes::Route<Self>] =
                &::ixc_core::routes::sort_routes([
                    #(#routes)*
                ]);
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

                    let publish_all = maybe_extract_attribute(imp)?;
                    for item in &mut imp.items {
                        match item {
                            syn::ImplItem::Fn(impl_fn) => {
                                let on_create = maybe_extract_attribute(impl_fn)?;
                                let publish = maybe_extract_attribute(impl_fn)?;
                                if publish.is_some() && on_create.is_some() {
                                    bail!("on_create and publish attributes must not be attached to the same function");
                                }
                                let publish = publish_all.clone().or(publish);
                                if publish.is_some() || on_create.is_some() { // TODO check visibility
                                    targets.push(PublishFn {
                                        signature: impl_fn.sig.clone(),
                                        on_create,
                                        publish,
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

#[derive(deluxe::ExtractAttributes, Clone, Debug)]
#[deluxe(attributes(publish))]
struct Publish {
    package: Option<String>,
    name: Option<String>,
}

#[derive(deluxe::ExtractAttributes, Debug)]
#[deluxe(attributes(on_create))]
struct OnCreate {
    message_name: Option<String>,
}

fn maybe_extract_attribute<T, R>(t: &mut T) -> manyhow::Result<Option<R>>
where
    T: deluxe::HasAttributes,
    R: deluxe::ExtractAttributes<T>,
{
    let mut have_attr = false;
    for attr in t.attrs() {
        if R::path_matches(attr.meta.path()) {
            have_attr = true;
        }
    }
    if !have_attr {
        return Ok(None);
    }
    Ok(Some(R::extract_attributes(t)?))
}

#[derive(Debug)]
struct PublishFn {
    signature: syn::Signature,
    on_create: Option<OnCreate>,
    publish: Option<Publish>,
}

/// This publishes a trait or struct impl block or a single fn within an impl block.
#[manyhow]
#[proc_macro_attribute]
pub fn publish(_attr: TokenStream2, item: TokenStream2) -> manyhow::Result<TokenStream2> {
    bail!("the #[publish] attribute is being used in the wrong context, possibly #[module_handler] or #[account_handler] has not been applied to the enclosing module")
}

/// This attribute macro should be attached to a trait that implements a handler API.
#[proc_macro_attribute]
pub fn handler_api(_attr: TokenStream, item: TokenStream) -> TokenStream {
    item
}
/// This attribute macro should be attached to the fn which is called when an account is created.
#[manyhow]
#[proc_macro_attribute]
pub fn on_create(_attr: TokenStream2, item: TokenStream2) -> manyhow::Result<TokenStream2> {
    bail!("the #[on_create] attribute is being used in the wrong context, possibly #[module_handler] or #[account_handler] has not been applied to the enclosing module")
}

/// Derive the `Resources` trait for a struct.
#[manyhow]
#[proc_macro_derive(Resources, attributes(state, client))]
pub fn derive_resources(input: DeriveInput) -> manyhow::Result<TokenStream2> {
    let name = input.ident;
    let mut str = match input.data {
        Data::Struct(str) => str,
        _ => bail!("can only derive Resources on structs"),
    };
    let mut field_inits = vec![];
    let mut prefix = 0u8;
    for field in str.fields.iter_mut() {
        let field_name = field.ident.as_ref().unwrap().clone();
        let ty = &field.ty.clone();
        if let Some(state) = maybe_extract_attribute::<_, State>(field)? {
            prefix = state.prefix.unwrap_or(prefix);
            field_inits.push(quote! {
                #field_name: <#ty as ::ixc_core::resource::StateObjectResource>::new(scope.state_scope, #prefix)?
            });
            prefix += 1;
        } else {
            bail!("only fields with #[state] attribute are supported currently");
        }
    }
    Ok(quote! {
        unsafe impl ::ixc_core::resource::Resources for #name {
            unsafe fn new(scope: &::ixc_core::resource::ResourceScope) -> ::core::result::Result<Self, ::ixc_core::resource::InitializationError> {
                Ok(Self {
                    #(#field_inits),*
                })
            }
        }
    })
}

#[derive(deluxe::ExtractAttributes, Debug)]
#[deluxe(attributes(state))]
struct State {
    prefix: Option<u8>,
    key: Option<Vec<Ident>>,
    value: Option<Vec<Ident>>,
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

/// Creates the message selector for the given message name.
#[proc_macro]
pub fn message_selector(item: TokenStream) -> TokenStream {
    let input_str = parse_macro_input!(item as LitStr);
    message_selector_from_str(&input_str.value()).into()
}

fn message_selector_from_str(msg: &str) -> TokenStream2 {
    let mut hasher = Blake2b512::new(); // TODO should we use 256 or 512?
    hasher.update(msg.as_bytes());
    let res = hasher.finalize();
    // take first 8 bytes and convert to u64
    let hash = u64::from_le_bytes(res[..8].try_into().unwrap());
    let expanded = quote! {
        #hash
    };
    expanded.into()
}