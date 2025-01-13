
      for node in nodes:
          node.start()

      for node in nodes:
          node.wait_for_unit("cunicu.service")

      import time
      time.sleep(10)

      print(alice.succeed("${lib.getExe config.nodes.alice.services.cunicu.package} --log-level debug5 --rpc-socket /run/cunicu/cunicu.sock status"))

      time.sleep(3600)
    '';
  }
)
{
  self,
  pkgs,
  lib,
}:
let
  common =
    { pkgs, config, lib, ... }:
    {
      imports = [ self.nixosModules.cunicu ];

      options = {
        ssh_port = lib.mkOption { type = lib.types.int; };
      };

      config = {
        networking.firewall.enable = false;

        services.openssh = {
          enable = true;
          settings = {
            PermitRootLogin = "yes";
            PermitEmptyPasswords = "yes";
          };
        };

        security.pam.services.sshd.allowNullPassword = true;
        virtualisation.forwardPorts = [
          {
            from = "host";
            host.port = config.ssh_port;
            guest.port = 22;
          }
        ];

        environment.systemPackages = [
            pkgs.evans
        ];

        system.stateVersion = "24.11";
      };
    };

  node =
    { config, pkgs, ... }:
    {
      imports = [ common ];

      services.cunicu.daemon = {
        enable = true;

        settings = {
          discover_peers = true;
          community = "nixos";
          backends = [ "grpc://signal:8080?insecure=true&skip_verify=true" ];
          ice.urls = [ "grpc://relay:8080?insecure=true&skip_verify=true" ];
          interfaces.wg0 = { };
        };
      };
    };

  staticAuthSecret = "some-not-so-secret";
in
pkgs.testers.runNixOSTest (
  { config, ... }:
  {
    name = "cunicu";

    meta = with pkgs.lib.maintainers; {
      maintainers = [ stv0g ];
    };

    nodes = {
      relay =
        {
          config,
          pkgs,
          lib,
          ...
        }:
        {
          imports = [ common ];

          ssh_port = 2220;

          services = {
            coturn = {
              enable = true;

              realm = "0l.de";

              use-auth-secret = true;
              static-auth-secret = staticAuthSecret;
            };

            cunicu.relay = {
              enable = true;
              urls = [
                "stun:relay"
                "turn:relay?secret=${staticAuthSecret}&ttl=1h"
                "turn:relay?secret=${staticAuthSecret}&ttl=1h&transport=tcp"
              ];
            };
          };
        };

      signal =
        {
          config,
          pkgs,
          lib,
          ...
        }:
        {
          imports = [ common ];

          ssh_port = 2221;

          services.cunicu.signal = {
            enable = true;
          };
        };

      alice = {
        imports = [ node ];

        ssh_port = 2222;
      };

      bob = {
        imports = [ node ];

        ssh_port = 2223;
      };
    };

    testScript = ''
      infra = [relay, signal]
      nodes = [alice, bob]
      all_nodes = nodes + infra

      for node in infra:
          node.start()

      relay.wait_for_unit("cunicu-relay.service")
      signal.wait_for_unit("cunicu-signal.service")

      for node in nodes:
          node.start()

      for node in nodes:
          node.wait_for_unit("cunicu.service")

      import time
      time.sleep(10)

      print(alice.succeed("${lib.getExe config.nodes.alice.services.cunicu.package} --log-level debug5 --rpc-socket /run/cunicu/cunicu.sock status"))

      time.sleep(3600)
    '';
  }
)