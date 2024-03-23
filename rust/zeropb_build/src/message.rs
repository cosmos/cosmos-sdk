use std::fmt::Write;

use prost_types::DescriptorProto;
use quote::{format_ident, quote};

use crate::ctx::Context;
use crate::field::gen_field;

pub(crate) fn gen_message(message: &DescriptorProto, ctx: &mut Context) -> anyhow::Result<()> {
    let name = format_ident!("{}", message.name.clone().unwrap());
    let mut fields = vec![];
    for field in message.field.iter() {
        let tokens = gen_field(field, ctx)?;
        fields.push(tokens);
    }

    ctx.add_item(quote!(
        #[repr(C)]
        pub struct #name {
            #(#fields)*
        }

    ))?;

    ctx.add_item(quote!(
        unsafe impl zeropb::ZeroCopy for #name {}
    ))?;

    // for nested in message.nested_type.iter() {
    //     gen_message(nested);
    // }
    // for enum_type in message.enum_type.iter() {
    //     gen_enum(enum_type);
    // }
    // for oneof in message.oneof_decl.iter() {
    //     gen_oneof(oneof);
    // }

    Ok(())

}
