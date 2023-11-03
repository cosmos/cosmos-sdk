use std::fmt::Write;
use prost_types::FileDescriptorProto;
use crate::ctx::Context;
use crate::message::gen_message;

pub(crate) fn gen_file(file: &FileDescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    for message in file.message_type.iter() {
        gen_message(message, ctx)?;
    }
    Ok(())
}