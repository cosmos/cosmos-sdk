use std::io::Result;

fn main() -> Result<()> {
    tonic_build::configure()
        .include_file("_includes.rs")
        .file_descriptor_set_path("src/file_descriptor_set.bin")
        .compile(
            &[
                "proto/example/counter/v1/tx.proto",
                "proto/example/counter/module/v1/module.proto"],
            &["proto/", "../../proto/"])?;
    Ok(())
}
