use crate::{Address, EventOf, Response};

pub struct PrepareContext<Events = ()> {}

impl<Events> PrepareContext<Events> {
    pub fn self_address(&self) -> Address {
        todo!()
    }

    pub fn caller(&self) -> Address {
        todo!()
    }

    pub fn new<NewEvents>(&self) -> PrepareContext<NewEvents> {
        todo!()
    }

    pub fn exec<In, Out, E>(self, f: impl FnMut(ExecContext<Events>, In) -> Response<Out>) -> AsyncResponse<In, Out, E>
    {
        todo!()
    }
}

pub type AsyncResponse<In, Out = (), E = String> = Result<AsyncResponseBody<In, Out>, E>;

pub struct AsyncResponseBody<In, Out> {}

impl<In, Out> AsyncResponseBody<In, Out> {
    pub fn exec(&self, ctx: ExecContext, in_param: In) -> Response<Out> {
        todo!()
    }
}

pub struct ExecContext<Events = ()> {}

impl<Events> ExecContext<Events> {
    pub fn ok<T, E>(&self, result: T) -> Response<T, E> {
        todo!()
    }

    pub fn self_address(&self) -> Address {
        todo!()
    }
}

impl<Events> ExecContext<Events> {
    pub fn new<NewEvents>(&self) -> ExecContext<NewEvents> {
        todo!()
    }

    pub fn emit<Event: EventOf<Events>, Err>(&self, event: Event) -> core::result::Result<(), Err> {
        todo!()
    }
}


pub struct Map<K, V> {}

impl<K, V> Map<K, V> {
    pub fn prepare_get(&self, ctx: PrepareContext, key: &K) -> AsyncResponse<(), V> {
        todo!()
    }

    pub fn prepare_set(&self, ctx: PrepareContext, key: &K) -> AsyncResponse<V> {
        todo!()
    }
}