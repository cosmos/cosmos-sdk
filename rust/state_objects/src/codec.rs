use interchain_schema::StructCodec;
use interchain_schema::types::*;
use interchain_schema::value::Value;

pub trait FieldTypes {}
impl FieldTypes for () {}
impl<A: Type> FieldTypes for (A,) {}
impl<A: Type, B: Type> FieldTypes for (A, B) {}
impl<A: Type, B: Type, C: Type> FieldTypes for (A, B, C) {}
impl<A: Type, B: Type, C: Type, D: Type> FieldTypes for (A, B, C, D) {}


pub trait ObjectValueField<'a> {
    type Value: Value<'a>;
}
pub trait ObjectKeyField<'a>: ObjectValueField<'a> {}

impl<'a, V: Value<'a>> ObjectValueField<'a> for V {
    type Value = V;
}
impl<'a> ObjectValueField<'a> for str {
    type Value = &'a str;
}

impl<'a> ObjectKeyField<'a> for str {}
impl <'a> ObjectKeyField<'a> for u8 {}
impl <'a> ObjectKeyField<'a> for u16 {}
impl <'a> ObjectKeyField<'a> for u32 {}
impl <'a> ObjectKeyField<'a> for u64 {}
impl <'a> ObjectKeyField<'a> for u128 {}
impl <'a> ObjectKeyField<'a> for i8 {}
impl <'a> ObjectKeyField<'a> for i16 {}
impl <'a> ObjectKeyField<'a> for i32 {}
impl <'a> ObjectKeyField<'a> for i64 {}
impl <'a> ObjectKeyField<'a> for i128 {}
impl <'a> ObjectKeyField<'a> for bool {}
impl <'a> ObjectKeyField<'a> for simple_time::Time {}
impl <'a> ObjectKeyField<'a> for simple_time::Duration {}
impl <'a> ObjectKeyField<'a> for interchain_core::Address {}

pub trait ObjectValue<'a> {
    type Value;
}

impl <'a> ObjectValue<'a> for () {
    type Value = ();
}
impl <'a, A: ObjectValueField<'a>> ObjectValue<'a> for A {
    type Value = A::Value;
}
impl <'a, A: ObjectValueField<'a>> ObjectValue<'a> for (A,) {
    type Value = (A::Value,);
}
impl <'a, A: ObjectValueField<'a>, B: ObjectValueField<'a>> ObjectValue<'a> for (A, B) {
    type Value = (A::Value, B::Value);
}
impl <'a, A: ObjectValueField<'a>, B: ObjectValueField<'a>, C: ObjectValueField<'a>> ObjectValue<'a> for (A, B, C) {
    type Value = (A::Value, B::Value, C::Value);
}
impl <'a, A: ObjectValueField<'a>, B: ObjectValueField<'a>, C: ObjectValueField<'a>, D: ObjectValueField<'a>> ObjectValue<'a> for (A, B, C, D) {
    type Value = (A::Value, B::Value, C::Value, D::Value);
}

pub trait ObjectKey<'a>: ObjectValue<'a> {}
impl <'a> ObjectKey<'a> for () {}
impl <'a, A: ObjectKeyField<'a>> ObjectKey<'a> for A {}
impl <'a, A: ObjectKeyField<'a>> ObjectKey<'a> for (A,) {}
impl <'a, A: ObjectKeyField<'a>, B: ObjectKeyField<'a>> ObjectKey<'a> for (A, B) {}
impl <'a, A: ObjectKeyField<'a>, B: ObjectKeyField<'a>, C: ObjectKeyField<'a>> ObjectKey<'a> for (A, B, C) {}
impl <'a, A: ObjectKeyField<'a>, B: ObjectKeyField<'a>, C: ObjectKeyField<'a>, D: ObjectKeyField<'a>> ObjectKey<'a> for (A, B, C, D) {}


// impl<T1: Type, A: ObjectValueField<T1>> ObjectValue<(T1,)> for A {
//     type Value<'a> = A::Value<'a>;
// }
// impl<T1: Type, A: ObjectValueField<T1>> ObjectKey<(T1,)> for A {}
// impl<T1: Type, A: ObjectValueField<T1>> ObjectValue for (A,) {
//     type Types = (T1,);
//     type Value<'a> = (A::Value<'a>,);
// }
// impl<A: ObjectKey> ObjectKey for A {}
// impl<A: ObjectKey> ObjectKey for (A,) {}
// impl<A: ObjectValue> ObjectValue for (A,) {
//     type Types = (A::Types,);
//     type Value<'a> = (A::Value<'a>,);
// }
// impl<A: ObjectKey> ObjectKey for (A,) {}
// impl<A: ObjectValue, B: ObjectValue> ObjectValue for (A, B) {
//     type Types = (A::Types, B::Types);
//     type Value<'a> = (A::Value<'a>, B::Value<'a>);
// }
// impl<A: ObjectKey, B: ObjectKey> ObjectKey for (A, B) {}
// impl<A: ObjectValue, B: ObjectValue, C: ObjectValue> ObjectValue for (A, B, C) {
//     type Types = (A::Types, B::Types, C::Types);
//     type Value<'a> = (A::Value<'a>, B::Value<'a>, C::Value<'a>);
// }
// impl<A: ObjectKey, B: ObjectKey, C: ObjectKey> ObjectKey for (A, B, C) {}
// impl<A: ObjectValue, B: ObjectValue, C: ObjectValue, D: ObjectValue> ObjectValue for (A, B, C, D) {
//     type Types = (A::Types, B::Types, C::Types, D::Types);
//     type Value<'a> = (A::Value<'a>, B::Value<'a>, C::Value<'a>, D::Value<'a>);
// }
// impl<A: ObjectKey, B: ObjectKey, C: ObjectKey, D: ObjectKey> ObjectKey for (A, B, C, D) {}
//
pub trait PrefixKey<'a, K: ObjectKey<'a>> {
    type Value;
}
// impl<K: ObjectKey> PrefixKey<K> for () {
//     type Value<'a> = ();
// }
// impl<K: ObjectKey> PrefixKey<K> for K {
//     type Value<'a> = K::Value<'a>;
// }
// impl<A: ObjectKeyField, B: ObjectKeyField> PrefixKey<(A, B)> for (A,) {
//     type Value<'a> = (A::Value<'a>,);
// }
// impl<A: ObjectKeyField, B: ObjectKeyField, C: ObjectKeyField> PrefixKey<(A, B, C)> for (A,) {
//     type Value<'a> = (A::Value<'a>,);
// }
// impl<A: ObjectKeyField, B: ObjectKeyField, C: ObjectKeyField> PrefixKey<(A, B, C)> for (A, B) {
//     type Value<'a> = (A::Value<'a>, B::Value<'a>);
// }
// impl<A: ObjectKeyField, B: ObjectKeyField, C: ObjectKeyField, D: ObjectKeyField> PrefixKey<(A, B, C, D)> for (A,) {
//     type Value<'a> = (A::Value<'a>,);
// }
// impl<A: ObjectKeyField, B: ObjectKeyField, C: ObjectKeyField, D: ObjectKeyField> PrefixKey<(A, B, C, D)> for (A, B) {
//     type Value<'a> = (A::Value<'a>, B::Value<'a>);
// }
// impl<A: ObjectKeyField, B: ObjectKeyField, C: ObjectKeyField, D: ObjectKeyField> PrefixKey<(A, B, C, D)> for (A, B, C) {
//     type Value<'a> = (A::Value<'a>, B::Value<'a>, C::Value<'a>);
// }
