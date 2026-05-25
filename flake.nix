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
            "-X phasionary/internal/version.Version=v${version}"
            "-X phasionary/internal/version.Commit=${commit}"
            "-X phasionary/internal/version.BuildDate=${buildDate}"
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

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.phasionary}/bin/phasionary";
      };
    };
}
