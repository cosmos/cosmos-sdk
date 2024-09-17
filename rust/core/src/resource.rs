//! Resource module.

/// An account or module handler's resources.
/// This is usually derived by the state management framework.
pub unsafe trait Resources {}

/// A resource is anything that an account or module can use to store its own
/// state or interact with other accounts and modules.
pub unsafe trait Resource: Sized {
    /// Creates a new resource.
    /// This should only be called in generated code.
    /// Do not call this function directly.
    unsafe fn new(initializer: &mut Initializer) -> Result<Self, InitializationError>;
}

/// A resource initializer.
pub struct Initializer {}

/// An error that occurs during resource initialization.
pub enum InitializationError {}

impl Initializer {
    /// Creates a new resource initializer.
    pub unsafe fn new() -> Self {
        Self {}
    }

    /// The current state prefix.
    pub unsafe fn state_prefix(&self) -> &[u8] {
        todo!()
    }

    /// Automatically generates a new state prefix for a state object.
    pub unsafe fn auto_state_prefix(&mut self) -> Result<&[u8], InitializationError> {
        todo!()
    }

    /// Reserves a new state prefix for a state object.
    pub unsafe fn reserve_state_prefix(&mut self, prefix: &[u8]) -> Result<&[u8], InitializationError> {
        todo!()
    }
}