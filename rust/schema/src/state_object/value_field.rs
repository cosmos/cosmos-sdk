use crate::types::ListElementType;
use crate::value::{ListElementValue, SchemaValue};

/// This trait describes value types that are to be used as generic parameters
/// where there is no lifetime parameter available.
/// Any types implementing this trait relate themselves to a type implementing [`SchemaValue`]
/// so that generic types taking a `Value` type parameter can use a borrowed value if possible.
pub trait ObjectFieldValue {
    /// The type that is used when inputting object values to functions.
    type In<'a>: SchemaValue<'a>;
    /// The type that is used in function return values.
    type Out<'a>: SchemaValue<'a>;
}

impl ObjectFieldValue for u8 {
    type In<'a> = u8;
    type Out<'a> = u8;
}
impl ObjectFieldValue for u16 {
    type In<'a> = u16;
    type Out<'a> = u16;
}
impl ObjectFieldValue for u32 {
    type In<'a> = u32;
    type Out<'a> = u32;
}
impl ObjectFieldValue for u64 {
    type In<'a> = u64;
    type Out<'a> = u64;
}
impl ObjectFieldValue for u128 {
    type In<'a> = u128;
    type Out<'a> = u128;
}
impl ObjectFieldValue for i8 {
    type In<'a> = i8;
    type Out<'a> = i8;
}
impl ObjectFieldValue for i16 {
    type In<'a> = i16;
    type Out<'a> = i16;
}
impl ObjectFieldValue for i32 {
    type In<'a> = i32;
    type Out<'a> = i32;
}
impl ObjectFieldValue for i64 {
    type In<'a> = i64;
    type Out<'a> = i64;
}
impl ObjectFieldValue for i128 {
    type In<'a> = i128;
    type Out<'a> = i128;
}
impl ObjectFieldValue for bool {
    type In<'a> = bool;
    type Out<'a> = bool;
}
/// This type is used to represent a borrowed `&str` in a state object.
pub struct Str;
impl ObjectFieldValue for Str {
    type In<'a> = &'a str;
    type Out<'a> = &'a str;
}
#[cfg(feature = "std")]
impl ObjectFieldValue for alloc::string::String {
    type In<'a> = &'a str;
    type Out<'a> = alloc::string::String;
}

/// This type is used to represent a borrowed `&[u8]` in a state object.
pub struct Bytes;
impl ObjectFieldValue for Bytes {
    type In<'a> = &'a [u8];
    type Out<'a> = &'a [u8];
}
impl ObjectFieldValue for simple_time::Time {
    type In<'a> = simple_time::Time;
    type Out<'a> = simple_time::Time;
}
impl ObjectFieldValue for simple_time::Duration {
    type In<'a> = simple_time::Duration;
    type Out<'a> = simple_time::Duration;
}
impl ObjectFieldValue for ixc_message_api::AccountID {
    type In<'a> = ixc_message_api::AccountID;
    type Out<'a> = ixc_message_api::AccountID;
}
impl<V: ObjectFieldValue> ObjectFieldValue for Option<V> {
    type In<'a> = Option<V::In<'a>>;
    type Out<'a> = Option<V::Out<'a>>;
}
impl<V: ObjectFieldValue> ObjectFieldValue for [V]
where
        for<'a> <V as ObjectFieldValue>::In<'a>: ListElementValue<'a>,
        for<'a> <<V as ObjectFieldValue>::In<'a> as SchemaValue<'a>>::Type: ListElementType,
        for<'a> <V as ObjectFieldValue>::Out<'a>: ListElementValue<'a>,
        for<'a> <<V as ObjectFieldValue>::Out<'a> as SchemaValue<'a>>::Type: ListElementType,
{
    type In<'a> = &'a [V::In<'a>];
    type Out<'a> = &'a [V::Out<'a>];
}
