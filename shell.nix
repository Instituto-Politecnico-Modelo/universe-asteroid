{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/release-23.11.tar.gz") { } }:

pkgs.mkShell {
  # nativeBuildInputs is usually what you want -- tools you need to run
  nativeBuildInputs = with pkgs; [
    docker-client
    gnumake
    ffmpeg

    # go development
    go
    go-outline
    gopls
    gopkgs
    go-tools
    delve
  ];

  hardeningDisable = [ "all" ];
}