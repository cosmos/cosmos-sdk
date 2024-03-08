use heck::ToSnakeCase;
use proc_macro2::Ident;
use prost_types::{FileDescriptorProto, MethodDescriptorProto, ServiceDescriptorProto};
use quote::{format_ident, quote};

use crate::ctx::{Context, TokenResult};
use crate::r#type::gen_message_name;

pub(crate) fn gen_service(
    fd: &prost_types::FileDescriptorProto,
    service: &ServiceDescriptorProto,
    ctx: &mut Context,
) -> anyhow::Result<()> {
    gen_service_server(fd, service, ctx)?;
    gen_service_client(fd, service, ctx)?;
    Ok(())
}

fn gen_service_server(fd: &FileDescriptorProto, service: &ServiceDescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    // check if the cosmos.msg.v1.Service option is true
    // if so all Server and Client methods are generated with &mut Context instead of &Context

    let name = format_ident!("{}Server", service.name.clone().unwrap());
    let mut methods = vec![];
    for method in service.method.iter() {
        let tokens = gen_server_method(method, ctx)?;
        methods.push(tokens);
    }

    ctx.add_item(quote!(
        pub trait #name {
            #(#methods)*
        }
    ))?;

    gen_server_impl(fd, service, name, ctx)
}

pub(crate) fn gen_server_method(
    method: &MethodDescriptorProto,
    ctx: &mut Context,
) -> TokenResult {
    let method_name = format_ident!("{}", method.name.clone().unwrap().to_snake_case());
    let input_type = method.input_type.clone().unwrap();
    let input_type_name = gen_message_name(&input_type)?;
    let output_type = method.output_type.clone().unwrap();
    let output_type_name = gen_message_name(&output_type)?;
    Ok(quote!(
        fn #method_name(&self, ctx: &mut ::zeropb::Context, req: &#input_type_name) -> ::zeropb::Result<#output_type_name>;
    ))
}

fn gen_server_impl(fd: &FileDescriptorProto, service: &ServiceDescriptorProto, name: Ident, ctx: &mut Context) -> anyhow::Result<()> {
    let mut matches = vec![];
    let mut i = 1u64;
    for method in service.method.iter() {
        let method_name = format_ident!("{}", method.name.clone().unwrap().to_snake_case());
        let req_type = method.input_type.clone().unwrap();
        let req_type = gen_message_name(&req_type)?;
        let match_arm = quote!(
            #i => self.#method_name(ctx, &*(req as *const #req_type)).map(|res| res.unsafe_unwrap()),
        );
        matches.push(match_arm);
        i += 1;
    }

    let package_name = fd.package.clone().unwrap();
    let full_name = format!("{}.{}", package_name, service.name.clone().unwrap());

    ctx.add_item(quote!(
        impl ::zeropb::Server for dyn #name
        {
            fn service_name(&self) -> &'static str {
                #full_name
            }

            fn route(&self, method_id: u64, ctx: &mut ::zeropb::Context, req: *mut u8, res: *mut *mut u8) -> ::zeropb::Code {
                unsafe {
                    let result: ::zeropb::RawResult<*mut u8> = match method_id {
                        #(#matches)*
                        _ => return ::zeropb::Code::Unimplemented,
                    };
                    match result {
                        Ok(ptr) => {
                            *res = ptr;
                            ::zeropb::Code::Ok
                        }
                        Err(err) => {
                            let ptr = err.msg.unsafe_unwrap();
                            if ptr != core::ptr::null_mut() {
                                *res = ptr;
                            }
                            err.code
                        }
                    }
                }
            }
        }
    ))
}

fn gen_service_client(fd: &FileDescriptorProto, service: &ServiceDescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    let svc_name = service.name.clone().unwrap();
    let client_name = format_ident!("{}Client", svc_name);
    ctx.add_item(quote!(
        pub struct #client_name {
            connection: zeropb::Connection,
            service_id: u64,
        }
    ))?;

    let mut methods = vec![];
    let mut i = 1;
    for method in service.method.iter() {
        let tokens = gen_client_method(i, method, ctx)?;
        i += 1;
        methods.push(tokens);
    }

    ctx.add_item(quote!(
        impl #client_name {
            #(#methods)*
        }
    ))?;

    let package_name = fd.package.clone().unwrap();
    let full_name = format!("{}.{}", package_name, svc_name);
    ctx.add_item(quote!(
        impl ::zeropb::Client for #client_name {
            fn service_name(&self) -> &'static str {
                #full_name
           }
        }
    ))
}

pub(crate) fn gen_client_method(
    i: u64,
    method: &MethodDescriptorProto,
    ctx: &mut Context,
) -> TokenResult {
    let method_name = format_ident!("{}", method.name.clone().unwrap().to_snake_case());
    let input_type = method.input_type.clone().unwrap();
    let input_type = gen_message_name(&input_type)?;
    let output_type = method.output_type.clone().unwrap();
    let output_type = gen_message_name(&output_type)?;

    Ok(quote!(
        pub fn #method_name(&self, ctx: &mut zeropb::Context, req: zeropb::Root<#input_type>) -> zeropb::Result<#output_type> {
            ::zeropb::connection_invoke(self.connection, #i, ctx, req)
        }
    ))
}
