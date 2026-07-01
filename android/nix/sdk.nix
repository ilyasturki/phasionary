# NixOS-patched Android SDK for building this app — prebuilt Google binaries
# don't run on NixOS, so nixpkgs patches them (notably aapt2). Build it once:
#
#   nix build --impure --expr "import $PWD/nix/sdk.nix" -o .nix-android-sdk
#
# Then point local.properties at it (the `just sdk` recipe does both). This is
# byte-for-byte the same derivation the sibling kilorep/android uses, so if you
# already built it there the store path is reused (no rebuild, no download).
#
# Build-only: build-tools 35 + platform 35 (matches compileSdk/targetSdk = 35).
# The emulator lives separately in the NixOS config as the `android-emulator`
# command.
let
  pkgs = import (builtins.getFlake "nixpkgs") {
    config = {
      allowUnfree = true; # Google's SDK components are unfree
      android_sdk.accept_license = true;
    };
  };
in
(pkgs.androidenv.composeAndroidPackages {
  platformVersions = [ "35" ];
  buildToolsVersions = [ "35.0.0" ];
}).androidsdk
