{ pkgs, lib, config, inputs, ... }:

let
  pkgs-unstable = import inputs.nixpkgs-unstable { system = pkgs.stdenv.system; };
in
{
 packages = [
   pkgs.git
   pkgs-unstable.goreleaser
 ];

  env.GREET = "devenv";
  # https://devenv.sh/languages/
  # languages.rust.enable = true;
  languages.go.enable = true;
  languages.python.enable = true;
  languages.python.uv.enable = true;

  # https://devenv.sh/processes/
  # processes.cargo-watch.exec = "cargo-watch";

  # https://devenv.sh/services/
  # services.postgres.enable = true;

  # https://devenv.sh/scripts/
  scripts.hello.exec = ''
    echo hello from $GREET
  '';

  enterShell = ''
    hello
    git --version
  '';

  # https://devenv.sh/tasks/
  # tasks = {
  #   "myproj:setup".exec = "mytool build";
  #   "devenv:enterShell".after = [ "myproj:setup" ];
  # };

  # https://devenv.sh/tests/
  enterTest = ''
    echo "Running tests"
    git --version | grep --color=auto "${pkgs.git.version}"
  '';

  # https://devenv.sh/pre-commit-hooks/
  # pre-commit.hooks.shellcheck.enable = true;
  #
  git-hooks.hooks = {
    gofmt = {
        enable = true;
      };
    govet = {
      enable = true;
      pass_filenames = false;
    };
    golangci-lint = {
      enable = true;
      pass_filenames = false;
    };
  };

  # See full reference at https://devenv.sh/reference/options/
}
