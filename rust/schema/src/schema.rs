use crate::oneof::OneOfType;
use crate::r#enum::EnumType;
use crate::structs::StructType;

pub enum SchemaType<'a> {
    Struct(StructType<'a>),
    Enum(EnumType<'a>),
    OneOf(OneOfType<'a>),
}