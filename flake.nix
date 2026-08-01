{
  description = "A modern ls with useful tables";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      version = "0.12.0";
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        nls = pkgs.buildGoModule {
          pname = "nls";
          inherit version;

          src = self;

          vendorHash = "sha256-BPSYn+oL6OTdKCcIcmTZnyUui+5IH+ZXD10FHjJlMOs=";

          subPackages = [ "cmd/nls" ];

          env.CGO_ENABLED = 0;

          ldflags = [
            "-s"
            "-w"
            "-X github.com/nolight132/nls/internal/version.Version=v${version}"
          ];

          nativeBuildInputs = [ pkgs.installShellFiles ];

          postInstall = ''
            $out/bin/nls --completion bash --completion zsh --completion fish
            installShellCompletion --cmd nls \
              --bash completion.bash \
              --zsh completion.zsh \
              --fish completion.fish
          '';

          meta = {
            description = "A modern ls with useful tables";
            homepage = "https://github.com/nolight132/nls";
            license = pkgs.lib.licenses.mit;
            mainProgram = "nls";
          };
        };

        default = nls;
      });

      apps = forAllSystems (pkgs: rec {
        nls = {
          type = "app";
          program = "${self.packages.${pkgs.stdenv.hostPlatform.system}.nls}/bin/nls";
        };
        default = nls;
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [ pkgs.go ];
        };
      });
    };
}
