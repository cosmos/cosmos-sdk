use crate::fields::FieldTypes;
use crate::structs::StructType;

pub(crate) const fn unnamed_struct_type<F: FieldTypes>() -> StructType<'static> {
    StructType {
        name: "",
        fields: F::FIELDS,
        sealed: false,
    }
}