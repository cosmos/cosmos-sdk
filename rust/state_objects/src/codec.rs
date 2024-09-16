use interchain_schema::StructCodec;
use interchain_schema::types::*;
use interchain_schema::value::Value;

pub trait FieldTypes {}
impl FieldTypes for () {}
impl<A: Type> FieldTypes for (A,) {}
impl<A: Type, B: Type> FieldTypes for (A, B) {}
impl<A: Type, B: Type, C: Type> FieldTypes for (A, B, C) {}
impl<A: Type, B: Type, C: Type, D: Type> FieldTypes for (A, B, C, D) {}


pub trait ObjectValueField<T: Type> {
    type Value<'a>: Value<'a, T>;
}
pub trait ObjectKeyField<T: Type>: ObjectValueField<T> {}

impl ObjectValueField<U8T> for u8 {
    type Value<'a> = u8;
}
impl ObjectKeyField<U8T> for u8 {}
impl ObjectValueField<U16T> for u16 {
    type Value<'a> = u16;
}
impl ObjectKeyField<U16T> for u16 {}
impl ObjectValueField<U32T> for u32 {
    type Value<'a> = u32;
}
impl ObjectKeyField<U32T> for u32 {}
impl ObjectValueField<U64T> for u64 {
    type Value<'a> = u64;
}
impl ObjectKeyField<U64T> for u64 {}
impl ObjectValueField<UIntNT<16>> for u128 {
    type Value<'a> = u128;
}
impl ObjectKeyField<UIntNT<16>> for u128 {}
impl ObjectValueField<I8T> for i8 {
    type Value<'a> = i8;
}
impl ObjectKeyField<I8T> for i8 {}
impl ObjectValueField<I16T> for i16 {
    type Value<'a> = i16;
}
impl ObjectKeyField<I16T> for i16 {}
impl ObjectValueField<I32T> for i32 {
    type Value<'a> = i32;
}
impl ObjectKeyField<I32T> for i32 {}
impl ObjectValueField<I64T> for i64 {
    type Value<'a> = i64;
}
impl ObjectKeyField<I64T> for i64 {}
impl ObjectValueField<IntNT<16>> for i128 {
    type Value<'a> = i128;
}
impl ObjectKeyField<IntNT<16>> for i128 {}
impl ObjectValueField<Bool> for bool {
    type Value<'a> = bool;
}
impl ObjectKeyField<Bool> for bool {}
impl ObjectValueField<StrT> for str {
    type Value<'a> = &'a str;
}
impl ObjectKeyField<StrT> for str {}
impl ObjectValueField<AddressT> for interchain_core::Address {
    type Value<'a> = interchain_core::Address;
}
impl ObjectKeyField<AddressT> for interchain_core::Address {}
impl ObjectValueField<TimeT> for simple_time::Time {
    type Value<'a> = simple_time::Time;
}
impl ObjectKeyField<TimeT> for simple_time::Time {}
impl ObjectValueField<DurationT> for simple_time::Duration {
    type Value<'a> = simple_time::Duration;
}

impl<T: Type, V: ObjectValueField<T>> ObjectValueField<NullableT<T>> for Option<V> {
    type Value<'a> = Option<V::Value<'a>>;
}
impl<S> ObjectValueField<StructT<S>> for S
where
        for<'a> S: StructCodec + 'a,
{
    type Value<'a> = S;
}

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
