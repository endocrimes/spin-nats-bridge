{
  description = "A NATS -> HTTP Proxy";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";

    nixpkgs-format.url = "github:nix-community/nixpkgs-fmt";
    nixpkgs-format.inputs.nixpkgs.follows = "nixpkgs";
    nixpkgs-format.inputs.flake-utils.follows = "flake-utils";

    gomod2nix.url = "github:nix-community/gomod2nix";
    gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
    gomod2nix.inputs.flake-utils.follows = "flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, nixpkgs-format, gomod2nix }:
  flake-utils.lib.eachDefaultSystem (system:
  let
    pkgs = import nixpkgs {
      inherit system;
    };

    defaultGo = pkgs.go_1_22;

    devPackages = with pkgs; [
      defaultGo

      gnumake
      git
      gotestsum
      golangci-lint
      natscli
      nats-server
      nixpkgs-format.defaultPackage.${system}
      gomod2nix.packages.${system}.default
    ];
  in {
    packages.default = gomod2nix.buildGoApplication {
      pname = "spin-nats-bridge";
      version = "0.1";
      pwd = ./.;
      src = ./.;
      modules = ./gomod2nix.toml;
      go = defaultGo;
    };

    devShells.default = pkgs.mkShell {
      buildInputs = devPackages;
    };
  });
}
