//! State object traits.

mod value;
mod key;
mod value_field;
mod key_field;
mod prefix;
mod field_types;

pub use value::ObjectValue;
pub use value_field::ObjectFieldValue;
pub use key_field::KeyFieldValue;
pub use prefix::PrefixKey;
pub use key::{ObjectKey, encode_object_key, decode_object_key};