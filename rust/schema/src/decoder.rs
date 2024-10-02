//! The decoder trait and error type.

use ixc_message_api::AccountID;
use crate::encoder::EncodeError;
use crate::list::ListDecodeVisitor;
use crate::mem::MemoryManager;
use crate::r#enum::EnumType;
use crate::structs::{StructDecodeVisitor, StructType};
use crate::value::SchemaValue;

/// The trait that decoders must implement.
pub trait Decoder<'a> {
    /// Decode a `u32`.
    fn decode_u32(&mut self) -> Result<u32, DecodeError>;
    /// Decode a `i32`.
    fn decode_i32(&mut self) -> Result<i32, DecodeError>;
    /// Decode a `u64`.
    fn decode_u64(&mut self) -> Result<u64, DecodeError>;
    /// Decode a `u128`.
    fn decode_u128(&mut self) -> Result<u128, DecodeError>;
    /// Decode a borrowed `str`.
    fn decode_borrowed_str(&mut self) -> Result<&'a str, DecodeError>;
    #[cfg(feature = "std")]
    /// Decode an owned `String`.
    fn decode_owned_str(&mut self) -> Result<alloc::string::String, DecodeError>;
    /// Decode a struct.
    fn decode_struct(&mut self, visitor: &mut dyn StructDecodeVisitor<'a>, struct_type: &StructType) -> Result<(), DecodeError>;
    /// Decode a list.
    fn decode_list(&mut self, visitor: &mut dyn ListDecodeVisitor<'a>) -> Result<(), DecodeError>;
    /// Decode an account ID.
    fn decode_account_id(&mut self) -> Result<AccountID, DecodeError>;
    /// Encode an enum value.
    fn decode_enum(&mut self, enum_type: &EnumType) -> Result<i32, DecodeError> {
        self.decode_i32()
    }
    /// Get the memory manager.
    fn mem_manager(&self) -> &'a MemoryManager;
}

/// Decode a single value.
pub fn decode<'a, D: Decoder<'a>, V: SchemaValue<'a>>(decoder: &mut D) -> Result<V, DecodeError> {
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