use std::io;
use std::io::{Read, Write};

use prost::Message;
use prost_types::compiler::{code_generator_response::File, CodeGeneratorRequest, CodeGeneratorResponse};

use zeropb_build::compile_fd;

fn main() -> io::Result<()> {
    let mut buf = Vec::new();
    io::stdin().read_to_end(&mut buf)?;

    let request = CodeGeneratorRequest::decode(buf.as_slice())?;
    let response = codegen(request).map_err(|e| io::Error::other(e))?;

    buf.clear();
    response.encode(&mut buf).expect("error encoding response");
    io::stdout().write_all(&buf)?;

    Ok(())
}

fn codegen(request: CodeGeneratorRequest) -> anyhow::Result<CodeGeneratorResponse> {
    let mut response = CodeGeneratorResponse::default();
    for file in request.proto_file {
        let (dest_path, contents) = compile_fd(&file)?;
        let mut f = File::default();
        f.name = Some(dest_path);
        f.content = Some(contents);
        response.file.push(f);
    }
    Ok(response)
}