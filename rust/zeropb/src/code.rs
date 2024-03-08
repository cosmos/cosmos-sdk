use cosmossdk_core::Code;
use crate::r#enum::ZeroCopyEnum;

unsafe impl ZeroCopyEnum for Code {
    const MAX_VALUE: u8 = 16;
}
