use std::fmt::Write;
use prost_types::FileDescriptorProto;
use crate::ctx::Context;
use crate::message::gen_message;
use crate::service::gen_service;

pub(crate) fn gen_file(file: &FileDescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    write!(ctx, "use zeropb;\n\n")?;
    for message in file.message_type.iter() {
        gen_message(message, ctx)?;
    }
    for service in file.service.iter() {
        gen_service(service, ctx)?;
    }
    Ok(())
}