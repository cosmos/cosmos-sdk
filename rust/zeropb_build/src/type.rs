use crate::ctx::{Context, TokenResult};
use prost_types::field_descriptor_proto::{Label, Type};
use prost_types::FieldDescriptorProto;
use std::fmt::Write;
use anyhow::anyhow;
use quote::{format_ident, quote};

pub(crate) fn gen_field_type(
    field: &FieldDescriptorProto,
    ctx: &mut Context,
) -> TokenResult {
    if field.proto3_optional() {
        todo!();
    }

    if field.label == Some(i32::from(Label::Repeated)) {
        match field.r#type() {
            Type::Group => {
                todo!()
            }
            Type::Message => {
                let typ = gen_message_full_name(&field.type_name.clone().ok_or(anyhow!("unexpected"))?)?;
                Ok(quote!(::zeropb::Repeated<#typ>))
            }
            Type::Enum => {
                todo!()
            }
            Type::Bytes => {
                Ok(quote!(::zeropb::Repeated<zeropb::Bytes>))
            }
            Type::String => {
                Ok(quote!(::zeropb::Repeated<zeropb::Str>))
            }
            ty => {
                let sty = gen_simple_type(ty, ctx)?;
                Ok(quote!(::zeropb::ScalarRepeated<#sty>))
            }
        }
    } else {
        match field.r#type() {
            Type::Group => {
                todo!()
            }
            Type::Message => {
                gen_message_full_name(&field.type_name.clone().ok_or(anyhow!("unexpected"))?)
            }
            Type::Enum => {
                todo!()
            }
            ty => {
                gen_simple_type(ty, ctx)
            }
        }
    }

}

pub(crate) fn gen_message_full_name(full_name: &str,) -> TokenResult {
    let name = full_name.split(".");
    // drop empty segments
    let name = name.filter(|s| !s.is_empty()).map(|s| format_ident!("{}", s)).collect::<Vec<_>>();
    // TODO: import package
    // let package = name.clone().take(name.clone().count() - 1).collect::<Vec<_>>();
    // ctx.header.push_str(&format!("use crate::{};\n", package.join("::")));
    Ok(quote!(crate::#(#name)::*))
}

pub(crate) fn gen_message_name(full_name: &str) -> TokenResult {
    let name = format_ident!("{}", full_name.split(".").last().ok_or(anyhow!("unexpected"))?);
    Ok(quote!(#name))
}

fn gen_simple_type(ty: Type, ctx: &mut Context) -> TokenResult {
    match ty {
        Type::String => {
            return Ok(quote!(::zeropb::Str));
        },
        Type::Bytes => {
            return Ok(quote!(::zeropb::Bytes));
        },
        _ => {
            return gen_scalar_type(ty, ctx);
        }
    }
}

fn gen_scalar_type(ty: Type, ctx: &mut Context) -> TokenResult {
    match ty {
        Type::Double => Ok(quote!(f64)),
        Type::Float => Ok(quote!(f32)),
        Type::Int64 | Type::Sfixed64 | Type::Sint64 => gen_i64(ctx),
        Type::Uint64 | Type::Fixed64 => gen_u64(ctx),
        Type::Int32 | Type::Sfixed32 | Type::Sint32 => gen_i32(ctx),
        Type::Uint32 | Type::Fixed32 => gen_u32(ctx),
        Type::Bool => Ok(quote!(bool)),
        _ => { return Err(anyhow::anyhow!("unexpected")); }
    }
}

fn gen_i32(ctx: &mut Context) -> TokenResult {
    if ctx.opts.handle_big_endian {
        Ok(quote!(rend::i32le))
    } else {
        Ok(quote!(i32))
    }
}

fn gen_u32(ctx: &mut Context) -> TokenResult {
    if ctx.opts.handle_big_endian {
        Ok(quote!(rend::u32le))
    } else {
        Ok(quote!(u32))
    }
}

fn gen_i64(ctx: &mut Context) -> TokenResult {
    if ctx.opts.handle_big_endian {
        Ok(quote!(rend::i64le))
    } else {
        Ok(quote!(i64))
    }
}

fn gen_u64(ctx: &mut Context) -> TokenResult {
    if ctx.opts.handle_big_endian {
        Ok(quote!(rend::u64le))
    } else {
        Ok(quote!(u64))
    }
}
