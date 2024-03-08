use prost_types::FieldDescriptorProto;

use crate::ctx::Context;
use crate::r#type::gen_field_type;
use std::fmt::Write;
use quote::{format_ident, quote};

pub(crate) fn gen_field(field: &FieldDescriptorProto, writer: &mut Context) -> anyhow::Result<proc_macro2::TokenStream> {
    let name = format_ident!("{}", field.name.clone().unwrap());
    let typ = gen_field_type(field, writer)?;

    Ok(quote!(
        pub #name: #typ,
    ))
}
