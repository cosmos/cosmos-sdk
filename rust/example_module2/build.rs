use std::io::Result;
fn main() -> Result<()> {
    prost_build::compile_protos(&["src/test1/test.proto", "src/cosmos/bank/v1beta1/tx.proto", "src/cosmos/base/v1beta1/coin.proto"], &["src/"])?;
    Ok(())
}