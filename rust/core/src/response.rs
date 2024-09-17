#![cfg_attr(feature = "try_trait_v2"), feature(try_trait_v2)]

use interchain_schema::value::Value;

/// Response is the type that should be used for message handler responses.
#[cfg(feature = "std")]
pub type Response<'a, R: ResponseValue, E: ResponseValue = ()> = Result<R::Owned, E::Owned>;

#[cfg(not(feature = "std"))]
pub struct Response<'a, R: ResponseValue, E: ResponseValue = ()> {
    _phantom: std::marker::PhantomData<&'a (R, E)>,
}

pub trait ResponseValue {
    type MaybeBorrowed<'a>;
    #[cfg(feature = "std")]
    type Owned;
}
impl ResponseValue for () {
    type MaybeBorrowed<'a> = ();
    #[cfg(feature = "std")]
    type Owned = ();
}
impl <V: Value> ResponseValue for V {
    type MaybeBorrowed<'a> = V::MaybeBorrowed<'a>;
    #[cfg(feature = "std")]
    type Owned = V::Owned;
}

#[cfg(feature = "try_trait_v2")]
impl <'a, R, E> core::ops::FromResidual for Response<'a, R, E> {
    fn from_residual(residual: <Self as Try>::Residual) -> Self {
        todo!()
    }
}

#[cfg(feature = "try_trait_v2")]
impl <'a, R, E> core::ops::Try for Response<'a, R, E> {
}
