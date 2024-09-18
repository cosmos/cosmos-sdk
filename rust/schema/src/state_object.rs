//! This crate defines traits specific to state objects within schemas.
use crate::value::{ResponseValue, Value};

// pub trait FieldTypes {}
// impl FieldTypes for () {}
// impl<A: Type> FieldTypes for (A,) {}
// impl<A: Type, B: Type> FieldTypes for (A, B) {}
// impl<A: Type, B: Type, C: Type> FieldTypes for (A, B, C) {}
// impl<A: Type, B: Type, C: Type, D: Type> FieldTypes for (A, B, C, D) {}

/// This trait is implemented for types that can be used as tuples of value fields in state objects.
pub trait ObjectValue {
    // type FieldType;
    /// The possibly borrowed value type to use.
    type Value<'a>;
}
impl ObjectValue for () {
    // type FieldTypes = ();
    type Value<'a> = ();
}
impl <A: Value> ObjectValue for A {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type,);
    type Value<'a> = A::MaybeBorrowed<'a>;
}
impl <A: Value> ObjectValue for (A,) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type,);
    type Value<'a> = (A::MaybeBorrowed<'a>);
}
impl <A: Value, B: Value> ObjectValue for (A,B) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type, B::MaybeBorrowed<'_>::Type);
    type Value<'a> = (A::MaybeBorrowed<'a>, B::MaybeBorrowed<'a>);
}
impl <A: Value, B: Value, C: Value> ObjectValue for (A,B,C) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type, B::MaybeBorrowed<'_>::Type, C::MaybeBorrowed<'_>::Type);
    type Value<'a> = (A::MaybeBorrowed<'a>, B::MaybeBorrowed<'a>, C::MaybeBorrowed<'a>);
}
impl <A: Value, B: Value, C: Value, D: Value> ObjectValue for (A,B,C,D) {
    // type FieldTypes = (A::MaybeBorrowed<'_>::Type, B::MaybeBorrowed<'_>::Type, C::MaybeBorrowed<'_>::Type, D::MaybeBorrowed<'_>::Type);
    type Value<'a> = (A::MaybeBorrowed<'a>, B::MaybeBorrowed<'a>, C::MaybeBorrowed<'a>, D::MaybeBorrowed<'a>);
}

/// This trait is implemented for types that can be used as key fields in state objects.
pub trait KeyFieldValue: Value {}
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
impl KeyFieldValue for interchain_message_api::Address {}

/// This trait is implemented for types that can be used as keys in state objects.
pub trait ObjectKey: ObjectValue {}
impl ObjectKey for () {}
impl <A: KeyFieldValue> ObjectKey for A {}
impl <A: KeyFieldValue> ObjectKey for (A,) {}
impl <A: KeyFieldValue, B: KeyFieldValue> ObjectKey for (A, B) {}
impl <A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue> ObjectKey for (A, B, C) {}
impl <A: KeyFieldValue, B: KeyFieldValue, C: KeyFieldValue, D: KeyFieldValue> ObjectKey for (A, B, C, D) {}

/// This trait is implemented for types that can be used as prefix keys in state objects.
pub trait PrefixKey<K: ObjectKey> {
    /// The possibly borrowed value type to use.
    type Value<'a>;
}
