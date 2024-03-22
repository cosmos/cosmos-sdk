{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs = { self, nixpkgs, flake-utils, ... }@inputs:
    {
      overlays.default = self: super: rec {
        simd = self.callPackage ./simapp { rev = self.shortRev or "dev"; };
        go = simd.go; # to build the tools (e.g. gomod2nix) using the same go version
        rocksdb = super.rocksdb.overrideAttrs (_: rec {
          version = "8.11.3";
          src = self.fetchFromGitHub {
            owner = "facebook";
            repo = "rocksdb";
            rev = "v${version}";
            sha256 = "sha256-OpEiMwGxZuxb9o3RQuSrwZMQGLhe9xLT1aa3HpI4KPs=";
          };
        });
      };
    } //
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          mkApp = drv: {
            type = "app";
            program = "${drv}/bin/${drv.meta.mainProgram}";
          };
          pkgs = import nixpkgs {
            inherit system;
            config = { };
            overlays = [
              inputs.gomod2nix.overlays.default
              self.overlays.default
            ];
          };
        in
        rec {
          packages = rec {
            default = simd;
            inherit (pkgs) simd;
          };
          apps = rec {
            default = simd;
            simd = mkApp pkgs.simd;
          };
          devShells = rec {
            default = with pkgs; mkShell {
              buildInputs = [
                simd.go
                rocksdb
                gomod2nix
              ];
            };
          };
          legacyPackages = pkgs;
        }
      ));
}
