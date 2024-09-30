//! This crate defines traits specific to state objects within schemas.

use crate::buffer::{Reader, Writer, WriterFactory};
use crate::decoder::{decode, DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::mem::MemoryManager;
use crate::types::ListElementType;
use crate::value::{ListElementValue, ResponseValue, Value};

// pub trait FieldTypes {}
// impl FieldTypes for () {}
// impl<A: Type> FieldTypes for (A,) {}
// impl<A: Type, B: Type> FieldTypes for (A, B) {}
// impl<A: Type, B: Type, C: Type> FieldTypes for (A, B, C) {}
// impl<A: Type, B: Type, C: Type, D: Type> FieldTypes for (A, B, C, D) {}

/// This trait is implemented for types that can be used as tuples of value fields in state objects.
pub trait ObjectValue {
    /// The type that is used when inputting object values to functions.
    type In<'a>;
    /// The type that is used in function return values.
    type Out<'a>;

    /// Encode each part of the value in reverse order.
    fn encode_reverse<'a, E: Encoder>(value: &Self::In<'a>, encoder: &mut E) -> Result<(), EncodeError>;

    /// Decode the value from the decoder.
    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Out<'a>, DecodeError>;
}
impl ObjectValue for () {
    // type FieldTypes = ();
    type In<'a> = ();
    type Out<'a> = ();

    fn encode_reverse<'a, E: Encoder>(_value: &Self::In<'a>, _encoder: &mut E) -> Result<(), EncodeError> {
        Ok(())
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Out<'a>, DecodeError> {
        Ok(())
    }
}
impl<A: ObjectFieldValue> ObjectValue for A {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type,);
    type In<'a> = A::In<'a>;
    type Out<'a> = A::Out<'a>;

    fn encode_reverse<'a, E: Encoder>(value: &Self::In<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <Self::In<'a> as Value<'a>>::encode(value, encoder)
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Out<'a>, DecodeError> {
        decode(decoder)
    }
}

impl<A: ObjectFieldValue> ObjectValue for (A,) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type,);
    type In<'a> = (A::In<'a>,);
    type Out<'a> = (A::Out<'a>,);

    fn encode_reverse<'a, E: Encoder>(value: &Self::In<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <A::In<'a> as Value<'a>>::encode(&value.0, encoder)
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Out<'a>, DecodeError> {
        Ok((decode(decoder)?,))
    }
}
impl<A: ObjectFieldValue, B: ObjectFieldValue> ObjectValue for (A, B) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type, B::MaybeBorrowed<'_>::Type);
    type In<'a> = (A::In<'a>, B::In<'a>);
    type Out<'a> = (A::Out<'a>, B::Out<'a>);

    fn encode_reverse<'a, E: Encoder>(value: &Self::In<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <B::In<'a> as Value<'a>>::encode(&value.1, encoder)?;
        <A::In<'a> as Value<'a>>::encode(&value.0, encoder)
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Out<'a>, DecodeError> {
        let a = decode(decoder)?;
        let b = decode(decoder)?;
        Ok((a, b))
    }
}
impl<A: ObjectFieldValue, B: ObjectFieldValue, C: ObjectFieldValue> ObjectValue for (A, B, C) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type, B::MaybeBorrowed<'_>::Type, C::MaybeBorrowed<'_>::Type);
    type In<'a> = (A::In<'a>, B::In<'a>, C::In<'a>);
    type Out<'a> = (A::Out<'a>, B::Out<'a>, C::Out<'a>);

    fn encode_reverse<'a, E: Encoder>(value: &Self::In<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <C::In<'a> as Value<'a>>::encode(&value.2, encoder)?;
        <B::In<'a> as Value<'a>>::encode(&value.1, encoder)?;
        <A::In<'a> as Value<'a>>::encode(&value.0, encoder)
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Out<'a>, DecodeError> {
        todo!()
    }
}
impl<A: ObjectFieldValue, B: ObjectFieldValue, C: ObjectFieldValue, D: ObjectFieldValue> ObjectValue for (A, B, C, D) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type, B::MaybeBorrowed<'_>::Type, C::MaybeBorrowed<'_>::Type, D::MaybeBorrowed<'_>::Type);
    type In<'a> = (A::In<'a>, B::In<'a>, C::In<'a>, D::In<'a>);
    type Out<'a> = (A::Out<'a>, B::Out<'a>, C::Out<'a>, D::Out<'a>);

    fn encode_reverse<'a, E: Encoder>(value: &Self::In<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <D::In<'a> as Value<'a>>::encode(&value.3, encoder)?;
        <C::In<'a> as Value<'a>>::encode(&value.2, encoder)?;
        <B::In<'a> as Value<'a>>::encode(&value.1, encoder)?;
        <A::In<'a> as Value<'a>>::encode(&value.0, encoder)
    }

    fn decode<'a, DEC: Decoder<'a>>(decoder: &mut DEC) -> Result<Self::Out<'a>, DecodeError> {
        todo!()
    }
}

/// This trait describes value types that are to be used as generic parameters
/// where there is no lifetime parameter available.
/// Any types implementing this trait relate themselves to a type implementing [`Value`]
/// so that generic types taking a `Value` type parameter can use a borrowed value if possible.
pub trait ObjectFieldValue {
    /// The type that is used when inputting object values to functions.
    type In<'a>: Value<'a>;
    /// The type that is used in function return values.
    type Out<'a>: Value<'a>;
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
impl ObjectFieldValue for str {
    type In<'a> = &'a str;
    type Out<'a> = &'a str;
}
#[cfg(feature = "std")]
impl ObjectFieldValue for alloc::string::String {
    type In<'a> = &'a str;
    type Out<'a> = alloc::string::String;
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
        for<'a> <<V as ObjectFieldValue>::In<'a> as Value<'a>>::Type: ListElementType,
        for<'a> <V as ObjectFieldValue>::Out<'a>: ListElementValue<'a>,
        for<'a> <<V as ObjectFieldValue>::Out<'a> as Value<'a>>::Type: ListElementType,
{
    type In<'a> = &'a [V::In<'a>];
    type Out<'a> = &'a [V::Out<'a>];
}


/// This trait is implemented for types that can be used as key fields in state objects.
pub trait KeyFieldValue: ObjectFieldValue {
    /// Encode the key segment as a non-terminal segment.
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        unimplemented!("encode")
    }

    /// Encode the key segment as the terminal segment.
    fn encode_terminal<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        Self::encode(key, writer)
    }

    /// Decode the key segment as a non-terminal segment.
    fn decode<'a, R: Reader<'a>>(reader: &mut R, memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        unimplemented!("decode")
    }

    /// Decode the key segment as the terminal segment.
    fn decode_terminal<'a, R: Reader<'a>>(reader: &mut R, memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        Self::decode(reader, memory_manager)
    }

    /// Get the size of the key segment as a non-terminal segment.
    fn out_size<'a>(key: &Self::In<'a>) -> usize {
        unimplemented!("size")
    }

    /// Get the size of the key segment as the terminal segment.
    fn out_size_terminal<'a>(key: &Self::In<'a>) -> usize {
        Self::out_size(key)
    }
}

impl KeyFieldValue for u8 {}
impl KeyFieldValue for u16 {}
impl KeyFieldValue for u32 {}
impl KeyFieldValue for u64 {}
impl KeyFieldValue for u128 {}
impl KeyFieldValue for i8 {}
impl KeyFieldValue for i16 {}
impl KeyFieldValue for i32 {}
impl KeyFieldValue for i64 {}
impl KeyFieldValue for i128 {}
impl KeyFieldValue for bool {}
impl KeyFieldValue for simple_time::Time {}
impl KeyFieldValue for simple_time::Duration {}
impl KeyFieldValue for ixc_message_api::AccountID {}

/// This trait is implemented for types that can be used as keys in state objects.
pub trait ObjectKey: ObjectValue {
    /// Encode the key.
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError>;

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError>;
}

impl ObjectKey for () {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        Ok(())
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        Ok(())
    }
}

impl<A: KeyFieldValue> ObjectKey for A {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode_terminal(key, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode_terminal(&mut reader, memory_manager)?;
        reader.done()?;
        Ok(a)
    }
}
impl<A: KeyFieldValue> ObjectKey for (A,) {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode(key.0, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        reader.done()?;
        Ok((a,))
    }
}

impl<A: KeyFieldValue, B: KeyFieldValue> ObjectKey for (A, B) {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode(key.0, writer)?;
        B::encode_terminal(key.1, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        let b = B::decode_terminal(&mut reader, memory_manager)?;
        reader.done()?;
        Ok((a, b))
    }
}

impl<A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue> ObjectKey for (A, B, C) {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode(key.0, writer)?;
        B::encode(key.1, writer)?;
        C::encode_terminal(key.2, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        let b = B::decode(&mut reader, memory_manager)?;
        let c = C::decode_terminal(&mut reader, memory_manager)?;
        reader.done()?;
        Ok((a, b, c))
    }
}

impl<A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue, D: KeyFieldValue> ObjectKey for (A, B, C, D) {
    fn encode<'a, W: Writer>(key: Self::In<'a>, writer: &mut W) -> Result<(), EncodeError> {
        A::encode(key.0, writer)?;
        B::encode(key.1, writer)?;
        C::encode(key.2, writer)?;
        D::encode_terminal(key.3, writer)
    }

    fn decode<'a>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<Self::Out<'a>, DecodeError> {
        let mut reader = input;
        let a = A::decode(&mut reader, memory_manager)?;
        let b = B::decode(&mut reader, memory_manager)?;
        let c = C::decode(&mut reader, memory_manager)?;
        let d = D::decode(&mut reader, memory_manager)?;
        reader.done()?;
        Ok((a, b, c, d))
    }
}

/// This trait is implemented for types that can be used as prefix keys in state objects.
pub trait PrefixKey<K: ObjectKey> {
    /// The possibly borrowed value type to use.
    type Value<'a>;
}
