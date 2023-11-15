use crate::ctx::Context;
use prost_types::field_descriptor_proto::{Label, Type};
use prost_types::FieldDescriptorProto;
use std::fmt::Write;

pub(crate) fn gen_field_type(
    field: &FieldDescriptorProto,
    ctx: &mut Context,
) -> anyhow::Result<()> {
    if field.proto3_optional() {
    }

    if field.label == Some(i32::from(Label::Repeated)) {
        match field.r#type() {
            Type::Group => {
                write!(ctx, "TODO")?;
            }
            Type::Message => {
                write!(ctx, "TODO")?;
            }
            Type::Enum => {
                write!(ctx, "TODO")?;
            }
            Type::Bytes => {
                write!(ctx, "zeropb::Repeated<zeropb::Bytes>")?;
            }
            Type::String => {
                write!(ctx, "zeropb::Repeated<zeropb::Str>")?;
            }
            ty => {
                write!(ctx, "zeropb::ScalarRepeated<")?;
                gen_simple_type(ty, ctx)?;
                write!(ctx, ">")?;
            }
        }
    }

    match field.r#type() {
        Type::Group => {
            write!(ctx, "TODO")?;
        }
        Type::Message => {
            write!(ctx, "TODO")?;
        }
        Type::Enum => {
            write!(ctx, "TODO")?;
        }
        ty => {
            gen_simple_type(ty, ctx)?;
        }
    }

    Ok(())
}

pub(crate) fn gen_message_type(full_name: &str, ctx: &mut Context) -> anyhow::Result<()> {
    // take last part of name after last dot
    let name = full_name.split(".").last().unwrap();
    write!(ctx, "{}", name)?;
    Ok(())
}

fn gen_simple_type(ty: Type, ctx: &mut Context) -> anyhow::Result<()> {
    match ty {
        Type::String => write!(ctx, "zeropb::Str")?,
        Type::Bytes => write!(ctx, "zeropb::Bytes")?,
        _ => {
            gen_scalar_type(ty, ctx)?;
        }
    }
    Ok(())
}

fn gen_scalar_type(ty: Type, ctx: &mut Context) -> anyhow::Result<()> {
    match ty {
        Type::Double => write!(ctx, "f64")?,
        Type::Float => write!(ctx, "f32")?,
        Type::Int64 | Type::Sfixed64 | Type::Sint64 => gen_i64(ctx)?,
        Type::Uint64 | Type::Fixed64 => gen_u64(ctx)?,
        Type::Int32 | Type::Sfixed32 | Type::Sint32 => gen_i32(ctx)?,
        Type::Uint32 | Type::Fixed32 => gen_u32(ctx)?,
        Type::Bool => write!(ctx, "bool")?,
        _ => {return Err(anyhow::anyhow!("unexpected"))}
    }
    Ok(())
}

fn gen_i32(ctx: &mut Context) -> anyhow::Result<()> {
    if ctx.opts.handle_big_endian {
        write!(ctx, "rend::i32le")?;
    } else {
        write!(ctx, "i32")?;
    }
    Ok(())
}

fn gen_u32(ctx: &mut Context) -> anyhow::Result<()> {
    if ctx.opts.handle_big_endian {
        write!(ctx, "rend::u32le")?;
    } else {
        write!(ctx, "u32")?;
    }
    Ok(())
}

fn gen_i64(ctx: &mut Context) -> anyhow::Result<()> {
    if ctx.opts.handle_big_endian {
        write!(ctx, "rend::i64le")?;
    } else {
        write!(ctx, "i64")?;
    }
    Ok(())
}

fn gen_u64(ctx: &mut Context) -> anyhow::Result<()> {
    if ctx.opts.handle_big_endian {
        write!(ctx, "rend::u64le")?;
    } else {
        write!(ctx, "u64")?;
    }
    Ok(())
}
