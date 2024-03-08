use crate::ctx::Context;
use crate::message::gen_message;
use crate::service::gen_service;
use prost_types::FileDescriptorProto;

pub(crate) fn gen_file(file: &FileDescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    for message in file.message_type.iter() {
        gen_message(message, ctx)?;
    }
    for service in file.service.iter() {
        gen_service(file, service, ctx)?;
    }
    Ok(())
}
