use std::io::Read;

use prost::Message;
use prost_types::FileDescriptorProto;
use std::borrow::Borrow;
use std::path::Path;
use std::{env, fs};
use syn::token::Star;

use crate::ctx::Context;
use crate::file::gen_file;

mod ctx;
mod field;
mod file;
mod message;
mod method;
mod opts;
mod service;
mod r#type;

#[cfg(test)]
mod tests {
    use std::borrow::Borrow;

    #[test]
    fn test1() {}
}

pub fn compile_fd_bytes(bz: &[u8]) -> std::io::Result<()> {
    // let mut gz = flate2::read::GzDecoder::new(bz);
    // let mut res = vec![];
    // res.reserve(0x10000);
    // gz.read_to_end(&mut res).unwrap();
    let fd = FileDescriptorProto::decode(bz.borrow()).unwrap();
    let (dest_path, contents) = compile_fd(&fd)?;
    let out_dir = env::var_os("OUT_DIR").unwrap();
    let dest_path = Path::new(&out_dir).join(dest_path);
    fs::create_dir_all(dest_path.parent().unwrap()).unwrap();
    fs::write(&dest_path, contents)?;
    Ok(())
}

pub fn compile_fd(fd: &FileDescriptorProto) -> std::io::Result<(String, String)> {
    let mut ctx = Context::default();
    gen_file(fd, &mut ctx).map_err(|e| std::io::Error::new(std::io::ErrorKind::Other, e))?;
    let contents = ctx.to_string();
    let dest_path = fd.name.as_ref().unwrap().replace(".proto", ".rs");
    Ok((dest_path, contents))
}