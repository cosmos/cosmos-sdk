use crate::response::Response;

/// An event bus that can be used to emit events.
pub struct EventBus<T> {
    _phantom: std::marker::PhantomData<T>,
}

impl<T> EventBus<T> {
    /// Emits an event to the event bus.
    pub fn emit<U>(&mut self, event: T) -> Response<()> {
        todo!()
    }
}