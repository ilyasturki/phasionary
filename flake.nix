{
  description = "Phasionary - Terminal-first project planning tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      packages.${system} = {
        phasionary = pkgs.buildGoModule {
          pname = "phasionary";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-tndx/Cjoc5Wm09xKiFR4LBFwQJONEhZkhyKPzsAYYbI=";
          ldflags = [ "-s" "-w" ];
          meta.platforms = pkgs.lib.platforms.linux;
        };
        default = self.packages.${system}.phasionary;
      };

      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [ go gopls gotools delve just ];
      };

      apps.${system}.default = {
        type = "app";
        program = "${self.packages.${system}.phasionary}/bin/phasionary";
      };
    };
}
