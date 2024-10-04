//! Routing system for message packets.

use allocator_api2::alloc::Allocator;
use ixc_message_api::handler::{HandlerError, HandlerErrorCode, HostBackend};
use ixc_message_api::packet::MessagePacket;

/// A router for message packets.
pub unsafe trait Router
where
    Self: 'static,
{
    /// The routes sorted by message selector.
    const SORTED_ROUTES: &'static [Route<Self>];
}

/// A route for a message packet.
pub type Route<T> = (u64, fn(&T, &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), HandlerError>);

/// Execute a message packet on a router.
pub fn exec_route<R: Router>(r: &R, packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), HandlerError> {
    let res = R::SORTED_ROUTES.binary_search_by_key(&packet.header().message_selector, |(selector, _)| *selector);
    match res {
        Ok(idx) => {
            R::SORTED_ROUTES[idx].1(r, packet, callbacks, allocator)
        }
        Err(_) => {
            Err(HandlerError::KnownCode(HandlerErrorCode::MessageNotHandled))
        }
    }
}

/// Sorts the routes by message selector.
pub const fn sort_routes<const N: usize, T: ?Sized>(mut arr: [Route<T>; N]) -> [Route<T>; N] {
    // const bubble sort
    loop {
        let mut swapped = false;
        let mut i = 1;
        while i < arr.len() {
            if arr[i - 1].0 > arr[i].0 {
                let left = arr[i - 1];
                let right = arr[i];
                arr[i - 1] = right;
                arr[i] = left;
                swapped = true;
            }
            i += 1;
        }
        if !swapped {
            break;
        }
    }
    arr
}

// TODO: can use https://docs.rs/array-concat/latest/array_concat/ to concat arrays then the above function to sort
// or https://docs.rs/constcat/latest/constcat/
