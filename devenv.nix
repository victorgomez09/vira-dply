{ pkgs, lib, config, inputs, ... }:


let
  pkgs-unstable = import inputs.nixpkgs-unstable { system = pkgs.stdenv.system; };
in

{
  env.GOOSE_DRIVER="postgres";
  env.PATH = "#{config.env.DEVENV_ROOT}/bin:$PATH";

  packages = [
    pkgs.air
    pkgs.goose
    pkgs.awscli2
    pkgs.golangci-lint
    pkgs.pkg-config
    pkgs.gpgme
    pkgs.btrfs-progs
  ];

  languages.javascript = {
    enable = true;
    package = pkgs-unstable.nodejs-slim_24;
    pnpm.enable = true;
    npm.enable = true;
  };


  languages.go.enable = true;
}
