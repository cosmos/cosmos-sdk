use impl_trait_for_tuples::impl_for_tuples;
use crate::field::Field;
use crate::structs::StructType;
use crate::types::{to_field, Type};

pub trait FieldTypes {
    const FIELDS: &'static [Field<'static>];
}

impl FieldTypes for () {
    const FIELDS: &'static [Field<'static>] = &[];
}

impl<A: Type> FieldTypes for (A,) {
    const FIELDS: &'static [Field<'static>] = &[
        to_field::<A>(),
    ];
}

impl<A: Type, B: Type> FieldTypes for (A, B) {
    const FIELDS: &'static [Field<'static>] = &[
        to_field::<A>(),
        to_field::<B>(),
    ];
}

impl<A: Type, B: Type, C: Type> FieldTypes for (A, B, C) {
    const FIELDS: &'static [Field<'static>] = &[
        to_field::<A>(),
        to_field::<B>(),
        to_field::<C>(),
    ];
}

impl<A: Type, B: Type, C: Type, D: Type> FieldTypes for (A, B, C, D) {
    const FIELDS: &'static [Field<'static>] = &[
        to_field::<A>(),
        to_field::<B>(),
        to_field::<C>(),
        to_field::<D>(),
    ];
}

pub(crate) const fn unnamed_struct_type<F: FieldTypes>() -> StructType<'static> {
    StructType {
        name: "",
        fields: F::FIELDS,
        sealed: false,
    }
}