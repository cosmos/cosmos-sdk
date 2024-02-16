use std::fmt::Write;

use heck::ToSnakeCase;
use prost_types::MethodDescriptorProto;

use crate::ctx::Context;
use crate::r#type::gen_message_name;
