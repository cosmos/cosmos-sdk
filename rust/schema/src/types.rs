//! This module defines the types that can be used in the schema at a type-level.
//!
//! Unless you are working with the implementation details of schema encoding, then you
//! should consider this module as something that ensures type safety.
//! This module uses a programming style known as type-level programming where types
//! are defined to build other types.
//! None of the types in this module are expected to be instantiated other than as type-level
//! parameters.

use crate::field::Field;
use crate::kind::Kind;

/// The `Type` trait is implemented for all types that can be used in the schema.
pub trait Type: Private {
    /// The kind of the type.
    const KIND: Kind;

    /// Whether the type is nullable.
    const NULLABLE: bool = false;

    /// The size limit of the type.
    const SIZE_LIMIT: Option<usize> = None;

    /// The element kind of a list type.
    const ELEMENT_KIND: Option<Kind> = None;

    /// The type that this type references, if any, otherwise ().
    type ReferencedType;
}
trait Private {}

/// The `U8T` type represents an unsigned 8-bit integer.
pub struct U8T;
impl Private for U8T {}
impl Type for U8T {
    const KIND: Kind = Kind::Uint8;
    type ReferencedType = ();
}
impl ListElementType for U8T {}

/// The `U16T` type represents an unsigned 16-bit integer.
pub struct U16T;
impl Private for U16T {}
impl Type for U16T {
    const KIND: Kind = Kind::Uint16;
    type ReferencedType = ();
}
impl ListElementType for U16T {}

/// The `U32` type represents an unsigned 32-bit integer.
pub struct U32T;
impl Private for U32T {}
impl Type for U32T {
    const KIND: Kind = Kind::Uint32;
    type ReferencedType = ();
}
impl ListElementType for U32T {}

/// The `U64T` type represents an unsigned 64-bit integer.
pub struct U64T;
impl Private for U64T {}
impl Type for U64T {
    const KIND: Kind = Kind::Uint64;
    type ReferencedType = ();
}

/// The `UIntNT` type represents an unsigned N-bit integer.
pub struct UIntNT<const N: usize> {}
impl<const N: usize> Private for UIntNT<N> {}
impl<const N: usize> Type for UIntNT<N> {
    const KIND: Kind = Kind::UIntN;
    const SIZE_LIMIT: Option<usize> = Some(N);
    type ReferencedType = ();
}

/// The `I8T` type represents a signed 8-bit integer.
pub struct I8T;
impl Private for I8T {}
impl Type for I8T {
    const KIND: Kind = Kind::Int8;
    type ReferencedType = ();
}

/// The `I16T` type represents a signed 16-bit integer.
pub struct I16T;
impl Private for I16T {}
impl Type for I16T {
    const KIND: Kind = Kind::Int16;
    type ReferencedType = ();
}

/// The `I32T` type represents a signed 32-bit integer.
pub struct I32T;
impl Private for I32T {}
impl Type for I32T {
    const KIND: Kind = Kind::Int32;
    type ReferencedType = ();
}

/// The `I64T` type represents a signed 64-bit integer.
pub struct I64T;
impl Private for I64T {}
impl Type for I64T {
    const KIND: Kind = Kind::Int64;
    type ReferencedType = ();
}

/// The `IntNT` type represents a signed integer represented by N bytes (not bits).
pub struct IntNT<const N: usize>;
impl<const N: usize> Private for IntNT<N> {}
impl<const N: usize> Type for IntNT<N> {
    const KIND: Kind = Kind::IntN;
    const SIZE_LIMIT: Option<usize> = Some(N);
    type ReferencedType = ();
}

/// The `Bool` type represents a boolean.
pub struct Bool;
impl Private for Bool {}
impl Type for Bool {
    const KIND: Kind = Kind::Bool;
    type ReferencedType = ();
}

/// The `StrT` type represents a string.
pub struct StrT;
impl Private for StrT {}
impl Type for StrT {
    const KIND: Kind = Kind::String;
    type ReferencedType = ();
}

/// The `AddressT` type represents an address.
pub struct AccountIDT;
impl Private for AccountIDT {}
impl Type for AccountIDT {
    const KIND: Kind = Kind::AccountID;
    type ReferencedType = ();
}

/// The `TimeT` type represents a time.
pub struct TimeT;
impl Private for TimeT {}
impl Type for TimeT {
    const KIND: Kind = Kind::Time;
    type ReferencedType = ();
}

/// The `DurationT` type represents a duration.
pub struct DurationT;
impl Private for DurationT {}
impl Type for DurationT {
    const KIND: Kind = Kind::Duration;
    type ReferencedType = ();
}

/// The `NullableT` type represents a nullable type.
pub struct NullableT<T> {
    _phantom: core::marker::PhantomData<T>,
}
impl <T> Private for NullableT<T> {}
impl <T: Type> Type for NullableT<T> {
    const NULLABLE: bool = true;
    const KIND: Kind = T::KIND;
    type ReferencedType = T::ReferencedType;
}

/// The `ListT` type represents a list type.
pub struct ListT<T: ListElementType> {
    _phantom: core::marker::PhantomData<T>,
}
impl <T:ListElementType> Private for ListT<T> {}
impl <T:ListElementType> Type for ListT<T> {
    const KIND: Kind = Kind::List;
    const ELEMENT_KIND: Option<Kind> = Some(T::KIND);
    type ReferencedType = T;
}

/// The `StructT` type represents a struct type.
pub struct StructT<T> {
    _phantom: core::marker::PhantomData<T>,
}
impl <T> Private for StructT<T> {}
impl <T> Type for StructT<T> {
    const KIND: Kind = Kind::Struct;
    type ReferencedType = T;
}
impl <T> ListElementType for StructT<T> {}

/// Represents a type that can be used as an element in a list.
pub(crate) trait ListElementType: Type {}

/// Converts a type to a field.
pub const fn to_field<T: Type>() -> Field<'static> {
    Field{
        name: "",
        kind: T::KIND,
        nullable: T::NULLABLE,
        referenced_type: "", // TODO
    }
}