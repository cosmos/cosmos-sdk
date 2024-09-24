use ixc_schema::state_object::ObjectValue;
use ixc_schema::value::{ResponseValue, AbstractValue};
use crate::error::ErrorMessage;

/// Response is the type that should be used for message handler responses.
#[cfg(feature = "std")]
// pub type Response<'a, R:ObjectValue, E: Value = ErrorMessage> =
//     Result<<<R as ObjectValue>::Value<'a> as ToOwned>::Owned, <<E as Value>::MaybeBorrowed<'a> as ToOwned>::Owned>;
pub type Response<'a, R, E = ErrorMessage> = core::result::Result<R, E>;

/// Response is the type that should be used for message handler responses.
#[cfg(not(feature = "std"))]
pub struct Response<'a, R: ResponseValue, E: ResponseValue = ()> {
    _phantom: std::marker::PhantomData<&'a (R, E)>,
}

#[cfg(feature = "try_trait_v2")]
impl<'a, R, E> core::ops::FromResidual for Response<'a, R, E> {
    fn from_residual(residual: <Self as Try>::Residual) -> Self {
        todo!()
    }
}

#[cfg(feature = "try_trait_v2")]
impl<'a, R, E> core::ops::Try for Response<'a, R, E> {}
