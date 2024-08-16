pub struct MessagePacket {
    header: MessagePacketHeader, // 64 + 64 + 32 + 8 + 128 + 64 + 664 = 1024
    data: [u8; 0xFC00], // 64512
}

pub struct MessagePacketHeader {
    address: Address, // 64
    caller: Address, // 64
    state_token: [u8; 32], //32
    gas_limit: u64, //8
    message_name: MessageName, //128
    params: [BufferRef; 4], // 16 * 4 = 64
    padding: [u8; 664] // 664
}

pub struct Address {
    len: u8,
    bytes: [u8; 63],
}

pub struct MessageName {
    len: u8,
    bytes: [u8; 127],
}

pub struct BufferRef {
    pointer: u64,
    capacity: u32,
    len: u32,
}
