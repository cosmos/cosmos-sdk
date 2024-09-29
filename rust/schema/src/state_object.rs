//! This crate defines traits specific to state objects within schemas.

use crate::decoder::{decode, DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::value::{ResponseValue, AbstractValue, Value};

// pub trait FieldTypes {}
// impl FieldTypes for () {}
// impl<A: Type> FieldTypes for (A,) {}
// impl<A: Type, B: Type> FieldTypes for (A, B) {}
// impl<A: Type, B: Type, C: Type> FieldTypes for (A, B, C) {}
// impl<A: Type, B: Type, C: Type, D: Type> FieldTypes for (A, B, C, D) {}

/// This trait is implemented for types that can be used as tuples of value fields in state objects.
pub trait ObjectValue {
    /// The possibly borrowed value type to use.
    type Value<'a>;

    /// Encode each part of the value in reverse order.
    fn encode_reverse<'a, E: Encoder>(value: &Self::Value<'a>, encoder: &mut E) -> Result<(), EncodeError>;

    /// Decode the value from the decoder.
    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Value<'a>, DecodeError>;
}
impl ObjectValue for () {
    // type FieldTypes = ();
    type Value<'a> = ();

    fn encode_reverse<'a, E: Encoder>(_value: &Self::Value<'a>, _encoder: &mut E) -> Result<(), EncodeError> {
        Ok(())
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Value<'a>, DecodeError> {
        Ok(())
    }
}
impl<A: AbstractValue> ObjectValue for A {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type,);
    type Value<'a> = A::Value<'a>;

    fn encode_reverse<'a, E: Encoder>(value: &Self::Value<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <Self::Value<'a> as Value<'a>>::encode(value, encoder)
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Value<'a>, DecodeError> {
        decode(decoder)
    }
}

impl<A: AbstractValue> ObjectValue for (A,) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type,);
    type Value<'a> = (A::Value<'a>,);

    fn encode_reverse<'a, E: Encoder>(value: &Self::Value<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <A::Value<'a> as Value<'a>>::encode(&value.0, encoder)
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Value<'a>, DecodeError> {
        Ok((decode(decoder)?,))
    }
}
impl<A: AbstractValue, B: AbstractValue> ObjectValue for (A, B) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type, B::MaybeBorrowed<'_>::Type);
    type Value<'a> = (A::Value<'a>, B::Value<'a>);

    fn encode_reverse<'a, E: Encoder>(value: &Self::Value<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <B::Value<'a> as Value<'a>>::encode(&value.1, encoder)?;
        <A::Value<'a> as Value<'a>>::encode(&value.0, encoder)
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Value<'a>, DecodeError> {
        let a = decode(decoder)?;
        let b = decode(decoder)?;
        Ok((a, b))
    }
}
impl<A: AbstractValue, B: AbstractValue, C: AbstractValue> ObjectValue for (A, B, C) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type, B::MaybeBorrowed<'_>::Type, C::MaybeBorrowed<'_>::Type);
    type Value<'a> = (A::Value<'a>, B::Value<'a>, C::Value<'a>);

    fn encode_reverse<'a, E: Encoder>(value: &Self::Value<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <C::Value<'a> as Value<'a>>::encode(&value.2, encoder)?;
        <B::Value<'a> as Value<'a>>::encode(&value.1, encoder)?;
        <A::Value<'a> as Value<'a>>::encode(&value.0, encoder)
    }

    fn decode<'a, D: Decoder<'a>>(decoder: &mut D) -> Result<Self::Value<'a>, DecodeError> {
        todo!()
    }
}
impl<A: AbstractValue, B: AbstractValue, C: AbstractValue, D: AbstractValue> ObjectValue for (A, B, C, D) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type, B::MaybeBorrowed<'_>::Type, C::MaybeBorrowed<'_>::Type, D::MaybeBorrowed<'_>::Type);
    type Value<'a> = (A::Value<'a>, B::Value<'a>, C::Value<'a>, D::Value<'a>);

    fn encode_reverse<'a, E: Encoder>(value: &Self::Value<'a>, encoder: &mut E) -> Result<(), EncodeError> {
        <D::Value<'a> as Value<'a>>::encode(&value.3, encoder)?;
        <C::Value<'a> as Value<'a>>::encode(&value.2, encoder)?;
        <B::Value<'a> as Value<'a>>::encode(&value.1, encoder)?;
        <A::Value<'a> as Value<'a>>::encode(&value.0, encoder)
    }

    fn decode<'a, DEC: Decoder<'a>>(decoder: &mut DEC) -> Result<Self::Value<'a>, DecodeError> {
        todo!()
    }
}

/// This trait is implemented for types that can be used as key fields in state objects.
pub trait KeyFieldValue: AbstractValue {}
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
#[cfg(feature = "address")]
impl KeyFieldValue for ixc_message_api::AccountID {}

/// This trait is implemented for types that can be used as keys in state objects.
pub trait ObjectKey: ObjectValue {}
impl ObjectKey for () {}
impl<A: KeyFieldValue> ObjectKey for A {}
impl<A: KeyFieldValue> ObjectKey for (A,) {}
impl<A: KeyFieldValue, B: KeyFieldValue> ObjectKey for (A, B) {}
impl<A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue> ObjectKey for (A, B, C) {}
impl<A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue, D: KeyFieldValue> ObjectKey for (A, B, C, D) {}

/// This trait is implemented for types that can be used as prefix keys in state objects.
pub trait PrefixKey<K: ObjectKey> {
    /// The possibly borrowed value type to use.
    type Value<'a>;
}
