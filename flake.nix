{
  description = "naisdevice";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    flake-root.url = "github:srid/flake-root";
    treefmt-nix.url = "github:numtide/treefmt-nix";
    devshell.url = "github:numtide/devshell";
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      imports = [
        inputs.flake-root.flakeModule
        inputs.treefmt-nix.flakeModule
        inputs.devshell.flakeModule
      ];
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];
      perSystem = {
        config,
        pkgs,
        system,
        ...
      }: {
        _module.args.pkgs = import inputs.nixpkgs {
          inherit system;
          overlays = [
            (import ./nix/overlay.nix)
          ];
        };
        devshells.default = {
          packages =
            (with pkgs; [
              gnumake
              go
              go-tools
              gopls
              gotools
              protobuf
            ])
            ++ [config.treefmt.build.wrapper];
        };
        treefmt.config = {
          inherit (config.flake-root) projectRootFile;
          package = pkgs.treefmt;

          programs = {
            alejandra.enable = true;
            deadnix.enable = true;
            gofumpt.enable = true;
            prettier.enable = true;
            statix.enable = true;
          };
        };
      };
    };
}
