use ixc_message_api::handler::{HandlerError, HandlerErrorCode};
use ixc_message_api::packet::MessagePacket;

pub unsafe trait Router
where
    Self: 'static,
{
    const SORTED_ROUTES: &'static [Route<Self>];
}

pub type Route<T> = (u64, fn(&T, &mut MessagePacket) -> Result<(), HandlerError>);

pub fn exec_route<R: Router>(r: &R, packet: &mut MessagePacket) -> Result<(), HandlerError> {
    let res = R::SORTED_ROUTES.binary_search_by_key(&packet.header().message_selector, |(selector, _)| *selector);
    match res {
        Ok(idx) => {
            R::SORTED_ROUTES[idx].1(r, packet)
        }
        Err(_) => {
            Err(HandlerError::KnownCode(HandlerErrorCode::MessageNotHandled))
        }
    }
}

pub const fn sort_routes<const N: usize, T>(mut arr: [Route<T>; N]) -> [Route<T>; N] {
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
