//! **WARNING: This is an API preview! Expect major bugs, glaring omissions, and breaking changes!**
//! This is a macro utility crate for ixc_core.

use proc_macro::{TokenStream};
use std::default::Default;
use proc_macro2::{Ident, TokenStream as TokenStream2};
use manyhow::{bail, ensure, manyhow};
use quote::{format_ident, quote, ToTokens};
use syn::{parse2, parse_macro_input, parse_quote, Attribute, Data, DeriveInput, Item, ItemMod, ItemTrait, LitStr, ReturnType, Signature, TraitItem, Type, Visibility};
use std::borrow::Borrow;
use blake2::{Blake2b512, Digest};
use deluxe::ExtractAttributes;
use heck::{ToUpperCamelCase};
use syn::punctuated::Punctuated;

#[derive(deluxe::ParseMetaItem)]
struct HandlerArgs(syn::Ident);

/// This derives an account handler.
#[manyhow]
#[proc_macro_attribute]
pub fn handler(attr: TokenStream2, mut item: ItemMod) -> manyhow::Result<TokenStream2> {
    let HandlerArgs(handler) = deluxe::parse2(attr)?;
    let items = &mut item.content.as_mut().unwrap().1;

    let mut publish_fns = vec![];
    let mut publish_traits = vec![];
    for item in items.iter_mut() {
        collect_publish_targets(&handler, item, &mut publish_fns, &mut publish_traits)?;
    }
    let mut builder = APIBuilder::default();
    for publish_target in publish_fns.iter() {
        derive_api_method(&handler, &quote! {#handler}, publish_target, &mut builder)?;
    }

    let client_ident = format_ident!("{}Client", handler);
    builder.define_client(&client_ident)?;
    builder.define_client_impl(&quote! {#client_ident}, &quote! {pub})?;
    builder.define_client_factory(&client_ident, &quote! {#handler})?;

    let on_create_msg = match builder.create_msg_name {
        Some(msg) => quote! {#msg},
        None => quote! {()},
    };
    let create_msg_lifetime = &builder.create_msg_lifetime;
    push_item(items, quote! {
        impl ::ixc::core::handler::Handler for #handler {
            const NAME: &'static str = stringify!(#handler);
            type Init<'a> = #on_create_msg #create_msg_lifetime;
        }
    })?;

    push_item(items, quote! {
        impl <'a> ::ixc::core::handler::InitMessage<'a> for #on_create_msg #create_msg_lifetime {
            type Codec = ::ixc::schema::binary::NativeBinaryCodec;
        }
    })?;

    let routes = &builder.routes;
    push_item(items, quote! {
        unsafe impl ::ixc::core::routing::Router for #handler {
            const SORTED_ROUTES: &'static [::ixc::core::routing::Route<Self>] =
                &::ixc::core::routing::sort_routes([
                    #(#routes)*
                ]);
        }
    })?;

    // TODO it would nice to be able to combine the routes rather than needing to check one by one
    let mut trait_routers = vec![];
    for publish_trait in publish_traits.iter() {
        let trait_ident = &publish_trait.ident;
        trait_routers.push(quote! {
            if let Some(rt) = ::ixc::core::routing::find_route::<dyn #trait_ident>(sel) {
                return rt.1(self, message_packet, callbacks, allocator)
            }
        })
    }

    push_item(items, quote! {
        impl ::ixc::message_api::handler::RawHandler for #handler {
            fn handle(&self, message_packet: &mut ::ixc::message_api::packet::MessagePacket, callbacks: &dyn ::ixc::message_api::handler::HostBackend, allocator: &dyn ::ixc::message_api::handler::Allocator) -> ::core::result::Result<(), ::ixc::message_api::code::ErrorCode> {
                let sel = message_packet.header().message_selector;
                if let Some(rt) = ::ixc::core::routing::find_route(sel) {
                    return rt.1(self, message_packet, callbacks, allocator)
                }

                #(#trait_routers)*

                Err(::ixc::message_api::code::ErrorCode::SystemCode(::ixc::message_api::code::SystemCode::MessageNotHandled))
            }
        }
    })?;

    push_item(items, quote! {
        impl ::ixc::core::handler::HandlerClient for #client_ident {
            type Handler = #handler;
        }
    })?;

    items.append(&mut builder.items);

    let expanded = quote! {
        #item
    };
    Ok(expanded)
}

fn push_item(items: &mut Vec<Item>, expanded: TokenStream2) -> manyhow::Result<()> {
    items.push(parse2::<Item>(expanded)?);
    Ok(())
}

fn collect_publish_targets(self_name: &syn::Ident, item: &mut Item, targets: &mut Vec<PublishFn>, traits: &mut Vec<PublishTrait>) -> manyhow::Result<()> {
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

                    let publish_all = maybe_extract_attribute(imp)?;

                    // TODO check for trait implementation
                    if imp.trait_.is_some() && publish_all.is_some() {
                        let trait_ident = imp.trait_.as_ref().unwrap().1.segments.first().unwrap().ident.clone();
                        traits.push(PublishTrait {
                            ident: trait_ident,
                        });
                        return Ok(());
                    }

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
                                        attrs: impl_fn.attrs.clone(),
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
    signature: Signature,
    on_create: Option<OnCreate>,
    publish: Option<Publish>,
    attrs: Vec<Attribute>,
}

#[derive(Debug)]
struct PublishTrait {
    ident: Ident,
}

/// This publishes a trait or struct impl block or a single fn within an impl block.
#[manyhow]
#[proc_macro_attribute]
pub fn publish(_attr: TokenStream2, item: TokenStream2) -> manyhow::Result<TokenStream2> {
    bail!("the #[publish] attribute is being used in the wrong context, possibly #[module_handler] or #[account_handler] has not been applied to the enclosing module")
}

/// This attribute macro should be attached to a trait that implements a handler API.
#[manyhow]
#[proc_macro_attribute]
pub fn handler_api(attr: TokenStream2, mut item_trait: ItemTrait) -> manyhow::Result<TokenStream2> {
    let mut builder = APIBuilder::default();
    let trait_ident = &item_trait.ident;
    let dyn_trait = quote!(dyn #trait_ident);
    for item in &item_trait.items {
        match item {
            TraitItem::Fn(f) => {
                let publish_target = PublishFn {
                    signature: f.sig.clone(),
                    on_create: None,
                    publish: None,
                    attrs: f.attrs.clone(),
                };
                derive_api_method(trait_ident, &dyn_trait, &publish_target, &mut builder)?;
            }
            _ => {}
        }
    }

    let client_trait_ident = format_ident!("{}Client", trait_ident);
    let client_impl_ident = format_ident!("{}Impl", client_trait_ident);
    builder.define_client(&client_impl_ident)?;
    builder.define_client_impl(&quote! {#client_trait_ident for #client_impl_ident}, &quote! {})?;
    builder.define_client_impl(&quote! {<T: ::ixc::core::handler::HandlerClient> #client_trait_ident for T
        where T::Handler: #trait_ident}, &quote! {})?;
    builder.define_client_factory(&client_impl_ident, &dyn_trait)?;
    builder.define_client_factory(&client_impl_ident, &quote! { #client_impl_ident})?;
    let items = &mut builder.items;
    let routes = &builder.routes;
    let client_signatures = &builder.client_signatures;
    Ok(quote! {
        #item_trait

        #(#items)*

        unsafe impl ::ixc::core::routing::Router for dyn #trait_ident {
            const SORTED_ROUTES: &'static [::ixc::core::routing::Route<Self>] =
                &::ixc::core::routing::sort_routes([
                    #(#routes)*
                ]);
        }

        impl ::ixc::message_api::handler::RawHandler for dyn #trait_ident {
            fn handle(&self, message_packet: &mut ::ixc::message_api::packet::MessagePacket, callbacks: &dyn ixc::message_api::handler::HostBackend, allocator: &dyn ::ixc::message_api::handler::Allocator) -> ::core::result::Result<(), ::ixc::message_api::code::ErrorCode> {
                ::ixc::core::routing::exec_route(self, message_packet, callbacks, allocator)
            }
        }

        pub trait #client_trait_ident {
            #( #client_signatures; )*
        }
    })
}

#[derive(Default)]
struct APIBuilder {
    items: Vec<Item>,
    routes: Vec<TokenStream2>,
    client_signatures: Vec<Signature>,
    client_methods: Vec<TokenStream2>,
    create_msg_name: Option<Ident>,
    create_msg_lifetime: TokenStream2,
}

impl APIBuilder {
    fn define_client(&mut self, client_ident: &Ident) -> manyhow::Result<()> {
        push_item(&mut self.items, quote! {
            pub struct #client_ident(::ixc::message_api::AccountID);
        })?;
        push_item(&mut self.items, quote! {
            impl ::ixc::core::handler::Client for #client_ident {
                fn new(account_id: ::ixc::message_api::AccountID) -> Self {
                    Self(account_id)
                }

                fn account_id(&self) -> ::ixc::message_api::AccountID {
                    self.0
                }
            }
        })
    }

    fn define_client_impl(&mut self, impl_target: &TokenStream2, visibility: &TokenStream2) -> manyhow::Result<()> {
        let client_methods = &self.client_methods;
        push_item(&mut self.items, quote! {
            impl #impl_target {
                #(#visibility #client_methods)*
            }
        })
    }


    fn define_client_factory(&mut self, client_ident: &Ident, factory_target: &TokenStream2) -> manyhow::Result<()> {
        push_item(&mut self.items, quote! {
            impl ::ixc::core::handler::Service for #factory_target {
                type Client = #client_ident;
            }
        })
    }
}

fn derive_api_method(handler_ident: &Ident, handler_ty: &TokenStream2, publish_target: &PublishFn, builder: &mut APIBuilder) -> manyhow::Result<()> {
    let signature = &publish_target.signature;
    let fn_name = &signature.ident;
    let ident_camel = fn_name.to_string().to_upper_camel_case();
    let msg_struct_name = format_ident!("{}{}", handler_ident, ident_camel);
    let mut signature = signature.clone();
    let mut new_inputs = Punctuated::new();
    let mut msg_fields = vec![];
    let mut msg_deconstruct = vec![];
    let mut fn_ctr_args = vec![];
    let mut msg_fields_init = vec![];
    let mut have_lifetimes = false;
    let mut context_name: Option<Ident> = None;
    for field in &signature.inputs {
        match field {
            syn::FnArg::Receiver(r) => {
                if r.mutability.is_some() {
                    bail!("error with fn {}: &self receiver on published fn's must be immutable", fn_name);
                }
            },
            syn::FnArg::Typed(pat_type) => {
                match pat_type.pat.as_ref() {
                    syn::Pat::Ident(ident) => {
                        let mut ty = pat_type.ty.clone();
                        match ty.as_mut() {
                            Type::Reference(tyref) => {
                                match tyref.elem.borrow() {
                                    Type::Path(path) => {
                                        if path.path.segments.first().unwrap().ident == "Context" {
                                            context_name = Some(ident.ident.clone());
                                            new_inputs.push(field.clone());
                                            continue;
                                        }

                                        if let Some(s) = path.path.segments.first() {
                                            if s.ident == "EventBus" {
                                                fn_ctr_args.push(quote! { &mut Default::default(), });
                                                continue;
                                            }
                                        }
                                    }
                                    _ => {}
                                }

                                have_lifetimes = true;
                                assert!(tyref.lifetime.is_none() ||
                                            tyref.lifetime.as_ref().unwrap().ident == "a"
                                        , "lifetime must be either unnamed or called 'a");
                                tyref.lifetime = Some(parse_quote!('a));
                            }
                            Type::Path(path) => {
                                if let Some(s) = path.path.segments.first() {
                                    if s.ident == "EventBus" {
                                        fn_ctr_args.push(quote! { Default::default(), });
                                        continue;
                                    }
                                }
                            }
                            _ => {}
                        }
                        msg_fields.push(quote! {
                                pub #ident: #ty,
                            });
                        msg_deconstruct.push(quote! {
                                #ident,
                            });
                        fn_ctr_args.push(quote! {
                                #ident,
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
        new_inputs.push(field.clone());
    }
    signature.inputs = new_inputs;
    let opt_lifetime = if have_lifetimes {
        quote! { <'a> }
    } else {
        quote! {}
    };
    let opt_underscore_lifetime = if have_lifetimes {
        quote! { <'_> }
    } else {
        quote! {}
    };

    push_item(&mut builder.items, quote! {
            #[derive(::ixc::SchemaValue)]
            #[sealed]
            pub struct #msg_struct_name #opt_lifetime {
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
        push_item(&mut builder.items, quote! {
                impl <'a> ::ixc::core::message::Message<'a> for #msg_struct_name #opt_lifetime {
                    const SELECTOR: ::ixc::message_api::header::MessageSelector = #selector;
                    type Response<'b> = <#return_type as ::ixc::core::message::ExtractResponseTypes>::Response;
                    type Error = <#return_type as ::ixc::core::message::ExtractResponseTypes>::Error;
                    type Codec = ::ixc::schema::binary::NativeBinaryCodec;
                }
            })?;
        ensure!(context_name.is_some(), "no context parameter found");
        let context_name = context_name.unwrap();
        builder.routes.push(quote! {
                    (< #msg_struct_name #opt_underscore_lifetime as ::ixc::core::message::Message >::SELECTOR, |h: &Self, packet, cb, a| {
                        unsafe {
                            let cdc = < #msg_struct_name as ::ixc::core::message::Message<'_> >::Codec::default();
                            let header = packet.header();
                            let in1 = header.in_pointer1.get(packet);
                            let mem = ::ixc::schema::mem::MemoryManager::new();
                            let #msg_struct_name { #(#msg_deconstruct)* } = ::ixc::schema::codec::decode_value::< #msg_struct_name >(&cdc, in1, &mem)?;
                            let mut ctx = ::ixc::core::Context::new_with_mem(header.account, header.caller, header.gas_left, cb, &mem);
                            let res = h.#fn_name(&mut ctx, #(#fn_ctr_args)*);
                            ::ixc::core::low_level::encode_response::< #msg_struct_name >(&cdc, res, a, packet)
                        }
                    }),
        });
        signature.output = parse_quote! {
            -> <#return_type as ::ixc::core::message::ExtractResponseTypes>::ClientResult
        };
        builder.client_signatures.push(signature.clone());
        builder.client_methods.push(quote! {
                #signature {
                    let msg = #msg_struct_name {
                        #(#msg_fields_init)*
                    };
                    unsafe { ::ixc::core::low_level::dynamic_invoke(#context_name, ::ixc::core::handler::Client::account_id(self), msg) }
                }
        });
    } else {
        builder.routes.push(quote! {
            (::ixc::core::account_api::ON_CREATE_SELECTOR, | h: &Self, packet, cb, a| {
                unsafe {
                    let cdc = < #msg_struct_name #opt_underscore_lifetime as::ixc::core::handler::InitMessage<'_> >::Codec::default();
                    let header = packet.header();
                    let in1 = header.in_pointer1.get(packet);
                    let mem = ::ixc::schema::mem::MemoryManager::new();
                    let #msg_struct_name { #(#msg_deconstruct)* } = ::ixc::schema::codec::decode_value::< #msg_struct_name > ( & cdc, in1, &mem)?;
                    let mut ctx =::ixc::core::Context::new_with_mem(header.account, header.caller, header.gas_left, cb, &mem);
                    let res = h.#fn_name(&mut ctx, #(#fn_ctr_args)*);
                    ::ixc::core::low_level::encode_default_response(res, a, packet)
                }
            }),}
        );
        builder.create_msg_name = Some(msg_struct_name);
        builder.create_msg_lifetime = opt_lifetime;
    }
    Ok(())
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
                #field_name: <#ty as ::ixc::core::resource::StateObjectResource>::new(scope.state_scope, #prefix)?
            });
            prefix += 1;
        }  else if let Some(client) = maybe_extract_attribute::<_, Client>(field)? {
            let account_id = client.0;
            field_inits.push(quote! {
                #field_name: <#ty as ::ixc::core::handler::Client>::new(::ixc::message_api::AccountID::new(#account_id))
            });
        } else {
            // TODO handle case where both #[state] and #[client] are present
            bail!("only fields with #[state] or #[client] attributes are supported currently");
        }
    }
    Ok(quote! {
        unsafe impl ::ixc::core::resource::Resources for #name {
            unsafe fn new(scope: &::ixc::core::resource::ResourceScope) -> ::core::result::Result<Self, ::ixc::core::resource::InitializationError> {
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
    #[deluxe(default)]
    key: Vec<Ident>,
    #[deluxe(default)]
    value: Vec<Ident>,
}

#[derive(deluxe::ExtractAttributes, Debug)]
#[deluxe(attributes(client))]
struct Client(u128);

// /// This attribute bundles account and module handlers into a package root which can be
// /// loaded into an application.
// #[proc_macro]
// pub fn package_root(item: TokenStream) -> TokenStream {
//     // let item = parse_macro_input!(item as File);
//     // let expanded = quote! {
//     //     #item
//     // };
//     // expanded.into()
//     TokenStream::default()
// }

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
