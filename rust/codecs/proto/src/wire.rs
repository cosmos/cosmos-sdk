use ixc_schema::encoder::EncodeError;
use ixc_schema::field::Field;
use ixc_schema::kind::Kind;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub(crate) enum WireType {
    Varint = 0,
    I64 = 1,
    LengthDelimited = 2,
    StartGroup = 3,
    EndGroup = 4,
    I32 = 5,
}

pub struct WireInfo {
    pub wire_type: WireType,
    pub unpacked: bool,
}

pub(crate) type FieldNumber = i32;

pub(crate) fn encode_tag(field_number: FieldNumber, wire_type: WireType) -> u64 {
    ((field_number as u64) << 3) | (wire_type as u64)
}

pub(crate) fn decode_tag(tag: u64) -> (FieldNumber, WireType) {
    ((tag >> 3) as FieldNumber, unsafe { core::mem::transmute((tag & 0b111) as u8) })
}

pub(crate) fn default_wire_info(field: &Field) -> Result<WireInfo, EncodeError> {
    let wire_type = match field.kind {
        Kind::Int8 | Kind::Uint8 | Kind::Int16 | Kind::Uint16 | Kind::Int32 | Kind::Uint32 | Kind::Bool | Kind::Enum |
        Kind::Int64 | Kind::Uint64 => WireType::Varint,
        Kind::String | Kind::Bytes | Kind::Struct => WireType::LengthDelimited,
        // encode as strings or byte arrays
        Kind::IntN | Kind::UIntN | Kind::Decimal => WireType::LengthDelimited,
        // encode as messages
        Kind::Time | Kind::Duration => WireType::LengthDelimited,
        Kind::List => {
            if let Some(kind) = field.element_kind {
                match kind {
                    Kind::Int8 | Kind::Uint8 | Kind::Int16 | Kind::Uint16 | Kind::Int32 | Kind::Uint32 | Kind::Bool |
                    Kind::Int64 | Kind::Uint64 => WireType::LengthDelimited,
                    _ => {
                        // a repeated field that isn't packed
                        let ty = default_wire_type_for_kind(kind)?;
                        return Ok(WireInfo {
                            wire_type: ty,
                            unpacked: true,
                        });
                    }
                }
            } else {
                return Err(EncodeError::UnknownError);
            }
        }
        _ => return Err(EncodeError::UnknownError),
    };
    Ok(WireInfo {
        wire_type,
        unpacked: false,
    })
}

pub(crate) fn default_wire_type_for_kind(kind: Kind) -> Result<WireType, EncodeError> {
    let t = match kind {
        Kind::Int8 | Kind::Uint8 | Kind::Int16 | Kind::Uint16 | Kind::Int32 | Kind::Uint32 | Kind::Bool | Kind::Enum |
        Kind::Int64 | Kind::Uint64 => WireType::Varint,
        Kind::String | Kind::Bytes | Kind::Struct => WireType::LengthDelimited,
        // encode as strings or byte arrays
        Kind::IntN | Kind::UIntN | Kind::Decimal => WireType::LengthDelimited,
        // encode as messages
        Kind::Time | Kind::Duration => WireType::LengthDelimited,
        Kind::List => return Err(EncodeError::UnknownError),
        _ => todo!()
    };
    Ok(t)
}