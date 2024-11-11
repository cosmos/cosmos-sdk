//! Resource module.

use ixc_message_api::AccountID;

/// An account or module handler's resources.
/// This is usually derived by the state management framework.
pub unsafe trait Resources: Sized {
    /// Initializes the resources.
    unsafe fn new(scope: &ResourceScope) -> Result<Self, InitializationError>;
}

/// The resource scope.
#[derive(Default)]
pub struct ResourceScope<'a> {
    /// The prefix of all state objects under this scope.
    pub state_scope: &'a [u8],

    /// The optional runtime account resolver.
    pub account_resolver: Option<&'a dyn AccountResolver>,
}

/// Resolves account names to account IDs.
pub trait AccountResolver {
    /// Resolves an account name to an account ID.
    fn resolve(&self, name: &str) -> Result<AccountID, InitializationError>;
}

#[cfg(feature = "std")]
impl AccountResolver for alloc::collections::BTreeMap<&str, AccountID> {
    fn resolve(&self, name: &str) -> Result<AccountID, InitializationError> {
        self.get(name).copied().ok_or(InitializationError::AccountNotFound)
    }
}

/// A resource is anything that an account or module can use to store its own
/// state or interact with other accounts and modules.
pub unsafe trait StateObjectResource: Sized {
    /// Creates a new resource.
    /// This should only be called in generated code.
    /// Do not call this function directly.
    unsafe fn new(scope: &[u8], prefix: u8) -> Result<Self, InitializationError>;
}

/// An error that occurs during resource initialization.
#[derive(Debug)]
pub enum InitializationError {
    /// An non-specific error occurred.
    Other,
    /// The account with the specified name could not be resolved.
    AccountNotFound,
}

impl<'a> ResourceScope<'a> {
    /// Resolves an account name to an account ID or returns a default account ID if provided.
    pub fn resolve_account(&self, name: &str, default: Option<AccountID>) -> core::result::Result<AccountID, InitializationError> {
        self.account_resolver
            .map(|resolver| resolver.resolve(name))
            .unwrap_or_else(|| default.ok_or(InitializationError::AccountNotFound))
    }
}