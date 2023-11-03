use std::fmt::Write;

use prost_types::DescriptorProto;

use crate::ctx::Context;
use crate::field::gen_field;

pub(crate) fn gen_message(message: &DescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    write!(ctx, "struct {} {{\n", message.name.clone().unwrap())?;

    for field in message.field.iter() {
        gen_field(field, ctx)?;
    }

    write!(ctx, "}}\n\n")?;
    write!(ctx, "unsafe impl zeropb::ZeroCopy for {} {{}}\n\n", message.name.clone().unwrap())?;
    Ok(())
    // for nested in message.nested_type.iter() {
    //     gen_message(nested);
    // }
    // for enum_type in message.enum_type.iter() {
    //     gen_enum(enum_type);
    // }
    // for oneof in message.oneof_decl.iter() {
    //     gen_oneof(oneof);
    // }
}
