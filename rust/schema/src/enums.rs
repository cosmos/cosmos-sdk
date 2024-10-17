use ixc_schema_macros::SchemaValue;
use crate::decoder::Decoder;
use crate::decoder::DecodeError;
use crate::encoder::{EncodeError, Encoder};
use crate::kind::Kind;
use crate::types::{ReferenceableType, Type};

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

pub unsafe trait EnumSchema:
ReferenceableType + TryFrom<Self::NumericType> + Into<Self::NumericType> + Clone
{
    const NAME: &'static str;
    const VALUES: &'static [EnumValueDefinition<'static>];
    const SEALED: bool;
    type NumericType: EnumNumericType;
    const ENUM_TYPE: EnumType<'static> = to_enum_type::<Self>();
}

pub const fn to_enum_type<E: EnumSchema>() -> EnumType<'static> {
    EnumType {
        name: E::NAME,
        values: E::VALUES,
        numeric_kind: E::NumericType::KIND,
        sealed: E::SEALED,
    }
}

trait EnumNumericType: Type {}
impl EnumNumericType for i32 {}
impl EnumNumericType for u16 {}
impl EnumNumericType for i16 {}
impl EnumNumericType for u8 {}
impl EnumNumericType for i8 {}

fn encode_enum<E: EnumSchema>(x: &E, encoder: &mut dyn Encoder)
                              -> Result<(), EncodeError>
where
    E::NumericType: Into<i32>,
{
    let value = encoder.encode_i32(E::into(x.clone()).into());
    value
}

fn decode_enum<E: EnumSchema>(decoder: &mut dyn Decoder) -> Result<E, DecodeError>
where
    E::NumericType: From<i32>,
{
    let x = decoder.decode_enum(&E::ENUM_TYPE)?;
    E::try_from(x.into())
        .map_err(|_| DecodeError::InvalidData)
}
