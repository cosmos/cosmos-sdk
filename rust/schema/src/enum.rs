use ixc_schema_macros::SchemaValue;
use crate::kind::Kind;

#[derive(Debug, Clone, Eq, PartialEq)]
#[non_exhaustive]
pub struct EnumType<'a> {
    pub name: &'a str,
    pub values: &'a [EnumValueDefinition<'a>],
    pub numeric_kind: Kind,
    pub sealed: bool,
}

#[derive(Debug, Clone, Default, PartialEq, Eq, SchemaValue)]
#[non_exhaustive]
pub struct EnumValueDefinition<'a> {
    pub name: &'a str,
    pub value: i32,
}

impl<'a> EnumValueDefinition<'a> {
    pub const fn new(name: &'a str, value: i32) -> Self {
        Self {
            name,
            value,
        }
    }
}
