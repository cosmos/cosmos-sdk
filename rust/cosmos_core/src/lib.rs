use cosmos_core_macros::service;

pub struct Context {}

impl Context {
    pub fn self_address(&self) -> Address {
        todo!()
    }

    pub fn sender(&self) -> Address {
        todo!()
    }
}

pub type Result<T> = core::result::Result<T, String>;

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
    padding: [u8; 664], // 664
}

#[derive(Clone, Copy, PartialEq, Eq)]
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

pub struct Time(u64);

#[service]
pub trait BlockService {
    fn current_time(&self, ctx: &Context) -> crate::Result<Time>;
}

pub trait Account {}

pub trait Module {}

pub trait Serializable {}

pub struct Item<T> {
    _phantom: core::marker::PhantomData<T>,
}

impl <T: Default> Item<T> {
    pub fn get(&self, ctx: &Context) -> Result<T> {
        todo!()
    }

    pub fn set(&self, ctx: &mut Context, value: &T) -> Result<()> {
        todo!()
    }
}

pub struct Map<K, V> {
    _phantom: core::marker::PhantomData<(K, V)>,
}

impl <K, V> Map<K, V> {
    pub fn get(&self, ctx: &Context, key: &K) -> Result<V> {
        todo!()
    }

    pub fn set(&self, ctx: &mut Context, key: &K, value: &V) -> Result<()> {
        todo!()
    }
}


pub trait Table {}

pub trait OnCreate {
    type InitMessage;

    fn on_create(&self, ctx: &mut Context, msg: &Self::InitMessage) -> Result<()>;
}