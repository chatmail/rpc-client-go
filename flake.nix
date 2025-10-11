{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs }: let
    supportedSystems = ["x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin"];
    forEachSupportedSystem = f:
      nixpkgs.lib.genAttrs supportedSystems (system:
        f {
          pkgs = import nixpkgs {
            inherit system;
          };
        });
  in {
    devShells = forEachSupportedSystem ({pkgs}: {
      default = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [
          go
          gopls
          deltachat-rpc-server
        ];
      };
      # not core included, for when you want to set it to a local core checkout via $PATH
      no_core = pkgs.mkShell {
        nativeBuildInputs = with pkgs; [
          go
          gopls
        ];
      };
    });
  };
}
