use imbl::{HashMap, OrdMap, Vector};

pub struct VersionedMultiStore {
    versions: Vector<MultiStore>,
}

pub struct MultiStore {
    stores: HashMap<u128, Store>,
}

pub enum Store {
    Unordered(HashMap<Vec<u8>, Vec<u8>>),
    Ordered(OrdMap<Vec<u8>, Vec<u8>>),
}

struct Update {
    store: u128,
    key: Vec<u8>,
    operation: Operation,
}

enum Operation {
    Set(Vec<u8>),
    Remove,
    Add(u128),
    SafeSub(u128),
    LazyAdd(u128),
}

type ChangeSet = Vec<Update>;