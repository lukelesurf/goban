{
  description = "Goban — self-hosted kanban board with MCP server";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          name = "goban";

          packages = with pkgs; [
            # Go toolchain
            go
            gopls
            gotools        # goimports, godoc, etc.
            golangci-lint

            # CGo deps (go-sqlite3 requires gcc + sqlite headers)
            gcc
            sqlite

            # Dev utilities
            delve          # Go debugger
          ];

          # CGo needs these
          CGO_ENABLED = "1";

          shellHook = ''
            echo "goban dev shell — go $(go version | awk '{print $3}')"
          '';
        };
      });
}
