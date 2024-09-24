use crate::decoder::{DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::kind::Kind;
use crate::types::{to_field, StructT, Type};
use crate::value::Value;

#[non_exhaustive]
#[derive(Debug, Clone)]
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

