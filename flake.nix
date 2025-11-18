{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: 
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        package = pkgs.buildGoModule {
          pname = "spaceage-tts";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;
          buildInputs = [];
        };
      in
      {
        packages = {
          default = package;
          spaceage-tts = package;
        };
      });
}
