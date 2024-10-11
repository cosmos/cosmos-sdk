//! State object traits.

mod value;
mod key;
mod value_field;
mod key_field;
mod prefix;
mod field_types;

pub use value::{ObjectValue, encode_object_value, decode_object_value};
pub use value_field::{ObjectFieldValue, Str, Bytes};
pub use key_field::KeyFieldValue;
pub use prefix::PrefixKey;
pub use key::{ObjectKey, encode_object_key, decode_object_key};
use crate::field::Field;

/// A type representing objects stored in key-value store state.
#[non_exhaustive]
#[derive(Debug, Clone, Eq, PartialEq)]
pub struct StateObjectType<'a> {
    /// The name of the object.
    pub name: &'a str,
    /// The fields that make up the primary key.
    pub key_fields: &'a [Field<'a>],
    /// The fields that make up the value.
    pub value_fields: &'a [Field<'a>],
    /// Whether to retain deletions in off-chain, indexed state.
    pub retain_deletions: bool,
}