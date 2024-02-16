use prost_types::FieldDescriptorProto;

use crate::ctx::Context;
use crate::r#type::gen_field_type;
use std::fmt::Write;

pub(crate) fn gen_field(field: &FieldDescriptorProto, writer: &mut Context) -> anyhow::Result<()> {
    write!(writer, "    ")?;
    write!(writer, "pub {}", field.name.clone().unwrap())?;
    write!(writer, " : ")?;
    gen_field_type(field, writer)?;
    writeln!(writer, ",")?;
    Ok(())
}
