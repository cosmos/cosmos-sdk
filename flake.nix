{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/master";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.utils.follows = "flake-utils";
    };
  };

  outputs = { self, nixpkgs, gomod2nix, flake-utils }:
    {
      overlays.default = pkgs: _: {
        simd = pkgs.callPackage ./simapp { rev = self.shortRev or "dev"; };
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
              gomod2nix.overlays.default
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
            default = simd;
            simd = with pkgs; mkShell {
              buildInputs = [
                go_1_21 # Use Go 1.21 version
                rocksdb
              ];
            };
          };
          legacyPackages = pkgs;
        }
      ));
}
