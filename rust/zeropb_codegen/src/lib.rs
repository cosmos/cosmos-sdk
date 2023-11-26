use std::io::Read;

use prost::Message;
use prost_types::FileDescriptorProto;
use std::borrow::Borrow;
use std::path::Path;
use std::{env, fs};

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

pub fn compile_fd(bz: &[u8]) -> std::io::Result<()> {
    // let mut gz = flate2::read::GzDecoder::new(bz);
    // let mut res = vec![];
    // res.reserve(0x10000);
    // gz.read_to_end(&mut res).unwrap();
    let fd = FileDescriptorProto::decode(bz.borrow()).unwrap();

    let mut ctx = Context::default();
    gen_file(&fd, &mut ctx).map_err(|e| std::io::Error::new(std::io::ErrorKind::Other, e))?;
    let out_dir = env::var_os("OUT_DIR").unwrap();
    let dest_path = Path::new(&out_dir).join(fd.name.unwrap().replace(".proto", ".rs"));
    fs::create_dir_all(dest_path.parent().unwrap()).unwrap();
    let contents = ctx.header + "\n" + &ctx.body;
    fs::write(&dest_path, contents)?;
    Ok(())
}
