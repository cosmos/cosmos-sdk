//! Well-known account mappings.

/// Well-known-accounts
include!(concat!(env!("OUT_DIR"), "/known_accounts.rs"));

/// Get the account ID for a given name.
pub const fn get_account_id(name: &str) -> Option<u128> {
    let mut left = 0;
    let mut right = KNOWN_ACCOUNTS.len();

    while left < right {
        let mid = left + (right - left) / 2;
        let (key, value) = KNOWN_ACCOUNTS[mid];

        match name.cmp(key) {
            core::cmp::Ordering::Equal => return Some(value),
            core::cmp::Ordering::Less => right = mid,
            core::cmp::Ordering::Greater => left = mid + 1,
        }
    }
    None
}

