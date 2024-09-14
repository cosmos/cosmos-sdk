pub unsafe trait Resource: Sized {
    unsafe fn new(initializer: &mut Initializer) -> Result<Self, InitializationError>;
}

pub struct Initializer {}

pub enum InitializationError {}

impl Initializer {
    pub fn new() -> Self {
        Self {}
    }

    pub fn state_prefix(&self) -> &[u8] {
        todo!()
    }

    pub fn auto_state_prefix(&mut self) -> Result<&[u8], InitializationError> {
        todo!()
    }

    pub fn reserve_state_prefix(&mut self, prefix: &[u8]) -> Result<&[u8], InitializationError> {
        todo!()
    }
}