use std::io::Result;
fn main() -> Result<()> {
    tonic_build::configure().compile(&["proto/example/counter/v1/tx.proto"], &["proto/"])?;
    Ok(())
}
