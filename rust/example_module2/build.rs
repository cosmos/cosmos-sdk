use std::io::Result;
fn main() -> Result<()> {
    prost_build::compile_protos(&["src/test.proto"], &["src/"])?;
    Ok(())
}