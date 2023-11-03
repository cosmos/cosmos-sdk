use std::io::Result;
fn main() -> Result<()> {
    zeropb_codegen::compile_protos()?;
    Ok(())
}