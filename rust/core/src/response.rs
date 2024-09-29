use bump_scope::Bump;
use ixc_message_api::packet::MessagePacket;
use ixc_schema::mem::MemoryManager;
use ixc_schema::value::{AbstractValue, ResponseValue};
use crate::error::Error;

/// Response is the type that should be used for message handler responses.
pub type Response<'a, R: ResponseValue, E = ()> = Result<ResponseBody<'a, R::Value<'a>>, Error<E>>;

/// ResponseBody is the type that should be used for message handler response bodies.
pub struct ResponseBody<'a, R: 'a> {
    _phantom: std::marker::PhantomData<&'a R>,
    bump: Bump,
    message_packet: MessagePacket,
    value: R,
}

impl <'a, R: 'a> ResponseBody<'a, R> {
    pub(crate) fn new(bump: Bump, message_packet: MessagePacket, value: R) -> Self {
        Self {
            _phantom: std::marker::PhantomData,
            bump,
            message_packet,
            value,
        }
    }
}

impl <'a, R: 'a> core::ops::Deref for ResponseBody<'a, R> {
    type Target = R;

    fn deref(&self) -> &Self::Target {
        &self.value
    }
}

// Response is the type that should be used for message handler responses.
// #[cfg(feature = "std")]
// pub type Response<'a, R:ObjectValue, E: Value = ErrorMessage> =
//     Result<<<R as ObjectValue>::Value<'a> as ToOwned>::Owned, <<E as Value>::MaybeBorrowed<'a> as ToOwned>::Owned>;
// pub type Response<'a, R, E = ErrorMessage> = core::result::Result<R, E>;

// /// Response is the type that should be used for message handler responses.
// pub struct Response<'a, R: ResponseValue, E: ResponseValue = ()> {
//     _phantom: std::marker::PhantomData<&'a (R, E)>,
//     memory_manager: MemoryManager<'a, 'a>,
//     message_packet: MessagePacket,
//     result: Result<(), ErrorCode>,
// }
//
// impl<'a, R: ResponseValue, E: ResponseValue> Response<'a, R, E> {
//     pub(crate) fn new(memory_manager: MemoryManager<'a, 'a>, message_packet: MessagePacket, result: Result<(), ErrorCode>) -> Self {
//         Self {
//             _phantom: std::marker::PhantomData,
//             memory_manager,
//             message_packet,
//             result,
//         }
//     }
// }
//
// #[cfg(feature = "try_trait_v2")]
// impl<'a, R: ResponseValue, E: ResponseValue> core::ops::FromResidual for Response<'a, R, E> {
//     fn from_residual(residual: <Self as core::ops::Try>::Residual) -> Self {
//         // Response::Err(residual)
//         todo!()
//     }
// }
//
// #[cfg(feature = "try_trait_v2")]
// impl<'a, R: ResponseValue, E: ResponseValue> core::ops::Try for Response<'a, R, E> {
//     type Output = R::Value<'a>;
//     type Residual = E::Value<'a>;
//
//     fn from_output(output: Self::Output) -> Self {
//         todo!()
//     }
//
//     fn branch(self) -> ControlFlow<Self::Residual, Self::Output> {
//         todo!()
//     }
// }
