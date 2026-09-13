{
  description = "Phasionary - Terminal-first project planning tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      version = pkgs.lib.fileContents ./VERSION;
      commit = self.shortRev or self.dirtyShortRev or "none";
      buildDate = self.lastModifiedDate or "unknown";

      phasionary-web = pkgs.buildNpmPackage {
        pname = "phasionary-web";
        inherit version;
        src = ./web;
        npmDepsHash = "sha256-BJcouDI6j+HxpYOwaOSqQbsxrD2j+uIGXtEyeIrW1S8=";
        doCheck = true;
        installPhase = ''
          runHook preInstall
          cp -r dist $out
          runHook postInstall
        '';
      };

      mkBinary =
        args:
        pkgs.buildGoModule (
          {
            inherit version;
            src = ./.;
            vendorHash = "sha256-XvjiBnzMxM/+ty9ZHT84FP7HO1h9cIxMbdorBJSIHCE=";
            ldflags = [
              "-s"
              "-w"
              "-X phasionary/internal/version.Version=v${version}"
              "-X phasionary/internal/version.Commit=${commit}"
              "-X phasionary/internal/version.BuildDate=${buildDate}"
            ];
            meta.platforms = pkgs.lib.platforms.linux;
          }
          // args
        );
    in
    {
      packages.${system} = {
        phasionary = mkBinary {
          pname = "phasionary";
          subPackages = [ "cmd/phasionary" ];
          nativeBuildInputs = [ pkgs.installShellFiles ];
          postInstall = ''
            installShellCompletion --cmd phasionary \
              --bash <($out/bin/phasionary completion bash) \
              --zsh <($out/bin/phasionary completion zsh) \
              --fish <($out/bin/phasionary completion fish)
          '';
        };
        inherit phasionary-web;
        phasionary-server = mkBinary {
          pname = "phasionary-server";
          subPackages = [ "cmd/phasionary-server" ];
          preBuild = ''
            rm -rf internal/webui/dist
            cp -r ${phasionary-web} internal/webui/dist
            chmod -R u+w internal/webui/dist
          '';
        };
        default = self.packages.${system}.phasionary;
      };

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.phasionary}/bin/phasionary";
      };

      nixosModules.phasionary-server = import ./nix/module.nix { inherit self; };
    };
}
