use interchain_schema::value::{ResponseValue};

/// Response is the type that should be used for message handler responses.
#[cfg(feature = "std")]
pub type Response<'a, R: ResponseValue, E: ResponseValue = ()> =
    Result<<<R as ResponseValue>::MaybeBorrowed<'a> as ToOwned>::Owned, <<E as ResponseValue>::MaybeBorrowed<'a> as ToOwned>::Owned>;

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
