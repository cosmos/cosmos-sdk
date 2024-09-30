use crate::decoder::{decode, DecodeError, Decoder};
use crate::encoder::{EncodeError, Encoder};
use crate::stateobject::value_field::ObjectFieldValue;
use crate::value::Value;

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

