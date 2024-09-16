use interchain_schema::StructCodec;
use interchain_schema::types::*;
use interchain_schema::value::Value;

pub trait FieldTypes {}
impl FieldTypes for () {}
impl<A: Type> FieldTypes for (A,) {}
impl<A: Type, B: Type> FieldTypes for (A, B) {}
impl<A: Type, B: Type, C: Type> FieldTypes for (A, B, C) {}
impl<A: Type, B: Type, C: Type, D: Type> FieldTypes for (A, B, C, D) {}


pub trait ObjectValueField {
    type Value<'a>: Value<'a>;
}
pub trait ObjectKeyField: ObjectValueField {}

impl ObjectValueField for u8 {
    type Value<'a> = u8;
}
impl ObjectKeyField for u8 {}
impl ObjectValueField for u16 {
    type Value<'a> = u16;
}
impl ObjectKeyField for u16 {}
impl ObjectValueField for u32 {
    type Value<'a> = u32;
}
impl ObjectKeyField for u32 {}
impl ObjectValueField for u64 {
    type Value<'a> = u64;
}
impl ObjectKeyField for u64 {}
impl ObjectValueField for u128 {
    type Value<'a> = u128;
}
impl ObjectKeyField for u128 {}
impl ObjectValueField for i8 {
    type Value<'a> = i8;
}
impl ObjectKeyField for i8 {}
impl ObjectValueField for i16 {
    type Value<'a> = i16;
}
impl ObjectKeyField for i16 {}
impl ObjectValueField for i32 {
    type Value<'a> = i32;
}
impl ObjectKeyField for i32 {}
impl ObjectValueField for i64 {
    type Value<'a> = i64;
}
impl ObjectKeyField for i64 {}
impl ObjectValueField for i128 {
    type Value<'a> = i128;
}
impl ObjectKeyField for i128 {}
impl ObjectValueField for bool {
    type Value<'a> = bool;
}
impl ObjectKeyField for bool {}
impl ObjectValueField for str {
    type Value<'a> = &'a str;
}
impl ObjectKeyField for str {}
impl ObjectValueField for simple_time::Time {
    type Value<'a> = simple_time::Time;
}
impl ObjectKeyField for simple_time::Time {}
impl ObjectValueField for simple_time::Duration {
    type Value<'a> = simple_time::Duration;
}
impl ObjectKeyField for simple_time::Duration {}
// impl<V> ObjectValueField for Option<V>
// {
//     type Value<'a> = Option<V>;
// }


pub trait ObjectValue {
    type Value<'a>;
}

pub trait ObjectKey: ObjectValue {}

// impl ObjectValue<()> for () {
//     type Value<'a> = ();
// }
// impl ObjectKey<()> for () {}
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
pub trait PrefixKey<K: ObjectKey> {
    type Value<'a>;
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
