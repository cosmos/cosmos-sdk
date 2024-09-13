pub trait Type: Private {}
trait Private {}

pub struct U8T;
pub struct U16T;
pub struct U32;
pub struct U64T;
pub struct U128T;
pub struct I8T;
pub struct I16T;
pub struct I32T;
pub struct I64T;
pub struct I128T;
pub struct Bool;
pub struct StrT;
pub struct AddressT;
pub struct NullableT<T> {
    _phantom: std::marker::PhantomData<T>,
}
pub struct ListT<T> {
    _phantom: std::marker::PhantomData<T>,
}
pub struct StructT<T> {
    _phantom: std::marker::PhantomData<T>,
}
