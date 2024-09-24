use crate::Address;
use crate::data_pointer::DataPointer;

pub const MESSAGE_HEADER_SIZE: usize = 512;

pub struct MessageHeader {
    pub target_account: u128, // 16 bytes
    pub sender_account: u128, // 16 bytes
    pub message_selector: [u8; 16],
    pub gas_limit: u64, // 8 bytes
    pub gas_consumed: u64, // 8 bytes
    pub in_pointer1: DataPointer, // 16 bytes
    pub in_pointer2: DataPointer, // 16 bytes
    pub out_pointer1: DataPointer, // 16 bytes
    pub out_pointer2: DataPointer, // 16 bytes
}

