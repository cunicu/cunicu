{ self, pkgs }:

pkgs.nixosTest {
  name = "cunicu";
  nodes.machine = { config, pkgs, ... }: {
    imports = [
      self.nixosModules.cunicu
    ];

    services.cunicu = {
      enable = true;
    };

    system.stateVersion = "24.11";
  };

  testScript = ''
    machine.wait_for_unit("cunicu.service")
  '';
}