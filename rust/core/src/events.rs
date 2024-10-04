use crate::Context;

/// An event bus that can be used to emit events.
pub struct EventBus<T> {
    _phantom: core::marker::PhantomData<T>,
}

impl <T> Default for EventBus<T> {
    fn default() -> Self {
        Self {
            _phantom: core::marker::PhantomData,
        }
    }
}


impl<T> EventBus<T> {
    /// Emits an event to the event bus.
    pub fn emit<U>(&mut self, ctx: &mut Context, event: T) -> crate::Result<()> {
        todo!()
    }
}