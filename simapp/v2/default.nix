{ lib
, buildGoApplication
, rocksdb
, stdenv
, static ? stdenv.hostPlatform.isStatic
, rev ? "dev"
}:

let
  pname = "simd";
  version = "v0.0.1";
  tags = [ "ledger" "netgo" "rocksdb" "grocksdb_no_link" ];
  ldflags = lib.concatStringsSep "\n" ([
    "-X github.com/cosmos/cosmos-sdk/version.Name=${pname}"
    "-X github.com/cosmos/cosmos-sdk/version.AppName=${pname}"
    "-X github.com/cosmos/cosmos-sdk/version.Version=${version}"
    "-X github.com/cosmos/cosmos-sdk/version.BuildTags=${lib.concatStringsSep "," tags}"
    "-X github.com/cosmos/cosmos-sdk/version.Commit=${rev}"
  ]);
in
buildGoApplication rec {
  inherit pname version ldflags tags;
  src = ./.;
  pwd = src;
  modules = ./gomod2nix.toml;
  subPackages = [ "simd" ];
  doCheck = false;
  buildInputs = [ rocksdb ];
  CGO_ENABLED = "1";
  CGO_LDFLAGS =
    if static then "-lrocksdb -pthread -lstdc++ -ldl -lzstd -lsnappy -llz4 -lbz2 -lz"
    else if stdenv.hostPlatform.isWindows then "-lrocksdb-shared"
    else "-lrocksdb -pthread -lstdc++ -ldl";

  postFixup = lib.optionalString stdenv.isDarwin ''
    ${stdenv.cc.targetPrefix}install_name_tool -change "@rpath/librocksdb.8.dylib" "${rocksdb}/lib/librocksdb.dylib" $out/bin/${pname}
  '';

  meta = with lib; {
    description = "example chain binary in cosmos-sdk repo";
    homepage = "https://github.com/cosmos/cosmos-sdk";
    license = licenses.asl20;
    mainProgram = pname + stdenv.hostPlatform.extensions.executable;
    platforms = platforms.all;
  };
}
