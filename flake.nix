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
      version = "0.2.0";
    in
    {
      packages.${system} = {
        phasionary = pkgs.buildGoModule {
          pname = "phasionary";
          inherit version;
          src = ./.;
          vendorHash = "sha256-tndx/Cjoc5Wm09xKiFR4LBFwQJONEhZkhyKPzsAYYbI=";
          ldflags = [
            "-s"
            "-w"
            "-X phasionary/internal/version.Version=${version}"
          ];

          nativeBuildInputs = [ pkgs.installShellFiles ];

          postInstall = ''
            installShellCompletion --cmd phasionary \
              --bash <($out/bin/phasionary completion bash) \
              --zsh <($out/bin/phasionary completion zsh) \
              --fish <($out/bin/phasionary completion fish)
          '';

          meta.platforms = pkgs.lib.platforms.linux;
        };
        default = self.packages.${system}.phasionary;
      };

      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          gopls
          gotools
          delve
          just
        ];
      };

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.phasionary}/bin/phasionary";
      };
    };
}
