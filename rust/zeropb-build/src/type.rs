use prost_types::field_descriptor_proto::Type;
use prost_types::FieldDescriptorProto;
use crate::ctx::Context;
use std::fmt::Write;

pub(crate) fn gen_type(field: &FieldDescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    match field.r#type() {
        Type::Double => { write!(ctx, "f64")? }
        Type::Float => { write!(ctx, "f32")? }
        Type::Int64 => { gen_i64(ctx)?; }
        Type::Uint64 => { gen_u64(ctx)?; }
        Type::Int32 => { gen_i32(ctx)?; }
        Type::Uint32 => { gen_u32(ctx)?; }
        Type::Fixed64 => { gen_u64(ctx)?; }
        Type::Fixed32 => { gen_u32(ctx)?; }
        Type::Sfixed32 => { gen_i32(ctx)?; }
        Type::Sfixed64 => { gen_i64(ctx)?; }
        Type::Sint32 => { gen_i32(ctx)?; }
        Type::Sint64 => { gen_i64(ctx)?; }
        Type::Bool => { write!(ctx, "bool")?; }
        Type::String => { write!(ctx, "zeropb::Str")?; }
        Type::Bytes => { write!(ctx, "zeropb::Bytes")?; }
        Type::Group => {
            write!(ctx, "TODO")?;
        }
        Type::Message => {
            write!(ctx, "TODO")?;
        }
        Type::Enum => {
            write!(ctx, "TODO")?;
        }
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
