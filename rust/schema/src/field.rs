use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::kind::Kind;
use crate::types::{to_field, StructT, Type};
use crate::value::SchemaValue;

#[non_exhaustive]
#[derive(Debug, Clone, Eq, PartialEq)]
pub struct Field<'a> {
    pub name: &'a str,
    pub kind: Kind,
    pub nullable: bool,
    pub referenced_type: &'a str,
}

impl <'a> Field<'a> {
    pub const fn with_name(mut self, name: &'a str) -> Self {
        self.name = name;
        self
    }
}

