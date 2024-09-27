//! The decoder trait and error type.
use crate::list::ListVisitor;
use crate::mem::MemoryManager;
use crate::structs::StructDecodeVisitor;
use crate::value::Value;

/// The trait that decoders must implement.
pub trait Decoder<'a> {
    /// Decode a `u32`.
    fn decode_u32(&mut self) -> Result<u32, DecodeError>;
    /// Decode a `u128`.
    fn decode_u128(&mut self) -> Result<u128, DecodeError>;
    /// Decode a borrowed `str`.
    fn decode_borrowed_str(&mut self) -> Result<&'a str, DecodeError>;
    #[cfg(feature = "std")]
    /// Decode an owned `String`.
    fn decode_owned_str(&mut self) -> Result<alloc::string::String, DecodeError>;
    /// Decode a struct.
    fn decode_struct<V: StructDecodeVisitor<'a>>(&mut self, visitor: &mut V) -> Result<(), DecodeError>;
    /// Decode a list.
    fn decode_list<T, V: ListVisitor<'a, T>>(&mut self, visitor: &mut V) -> Result<(), DecodeError>;
    /// Get the memory manager.
    fn mem_manager(&self) -> &MemoryManager<'a, 'a>;
}

/// Decode a single value.
pub fn decode<'a, D: Decoder<'a>, V: Value<'a>>(decoder: &mut D) -> Result<V, DecodeError> {
    let mut state = V::DecodeState::default();
    V::visit_decode_state(&mut state, decoder)?;
    V::finish_decode_state(state, decoder.mem_manager())
}

/// A decoding error.
#[derive(Debug)]
pub enum DecodeError {
    /// The input data is out of data.
    OutOfData,
    /// The input data is invalid.
    InvalidData,
    /// An unknown and unhandled field number was encountered.
    UnknownFieldNumber
}

// pub trait DecodeHelper<'a>: Default {
//     type Value;
//     type MemoryHandle;
//
//     fn finish(self) -> (Self::Value, Option(Self::MemoryHandle));
// }
//
// impl<'a> DecodeHelper<'a> for i32 {
//     type Value = i32;
//     type MemoryHandle = ();
//
//     fn finish(self) -> (Self::Value, Option(Self::MemoryHandle)) {
//         (self, None)
//     }
// }
//
// #[derive(Default)]
// pub struct BorrowedStrHelper<'a> {
//     pub(crate) s: &'a str,
//     pub(crate) owner: Option<BumpString<'a, 'a>>,
// }
//
// impl<'a> DecodeHelper<'a> for BorrowedStrHelper<'a> {
//     type Value = &'a str;
//     type MemoryHandle = Option<BumpString<'a, 'a>>;
//
//     fn finish(self) -> (Self::Value, Self::MemoryHandle) {
//         (self.s, self.owner)
//     }
// }
//
//
// pub struct SliceHelper<'a, T: ArgValue<'a>> {
//     // TODO maybe there's a way that the underlying data could already be a slice so we can just borrow:
//     // pub(crate) s: &'a [T],
//     // TODO why not MutBumpVec?
//     pub(crate) vec: Option<BumpVec<'a, 'a, T>>,
//     pub(crate) helpers: BumpVec<'a, 'a, T::DecodeState::MemoryHandle>,
// }
//
// impl<'a, T: ArgValue<'a>> Default for SliceHelper<'a, T> {
//     fn default() -> Self {
//         Self {
//             vec: None,
//         }
//     }
// }
//
// impl<'a, T: ArgValue<'a>> DecodeHelper<'a> for SliceHelper<'a, T> {
//     type Value = &'a [T];
//     type MemoryHandle = (BumpVec<'a, 'a, T>, BumpVec<'a, 'a, T::DecodeState::MemoryHandle>);
//
//     fn finish(self) -> (Self::Value, Self::MemoryHandle) {
//         todo!()
//     }
// }
//
// impl<'a, T: ArgValue<'a>> ListVisitor<'a, T> for SliceHelper<'a, T> {
//     fn init(&mut self, len: usize, scope: &'a mut BumpScope<'a>) -> Result<(), DecodeError> {
//         let mut vec = BumpVec::new_in(scope);
//         vec.reserve(len);
//         Ok(())
//     }
//
//     fn next<D: Decoder<'a>>(&mut self, decoder: &mut D) -> Result<(), DecodeError> {
//         let vec = if let Some(vec) = &mut self.vec {
//             vec
//         } else {
//             let mut vec= BumpVec::new_in(decoder.scope());
//             self.vec = Some(vec);
//             self.vec.as_mut().unwrap()
//         };
//         let helper: T::DecodeState = Default::default();
//         helper.decode(decoder)?;
//         let (value, memory_handle) = helper.finish();
//         vec.push(value);
//     }
// }