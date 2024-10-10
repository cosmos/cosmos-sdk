use alloy_sol_types::abi::encode;
use alloy_sol_types::SolValue;
use ixc_schema::codec::ValueEncodeVisitor;
use ixc_schema::encoder::EncodeError;
use ixc_schema::list::ListEncodeVisitor;
use ixc_schema::structs::StructEncodeVisitor;
use ixc_schema::value::SchemaValue;

struct Encoder {}
