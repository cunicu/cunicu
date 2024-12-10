{
  pkgs,
  lib,
	config,
	...
}:
let
  cfg = config.services.cunicu;
in
{
  options.services.cunicu = {
    enable = lib.mkEnableOption "Enable cunicu VPN client";

    settings = lib.mkOption {
      type = lib.types.anything;
      description = ''
        cunicu configuration

        See: https://cunicu.li/docs/config
      '';
			default = {};
      
    };
  };

  config = lib.mkIf cfg.enable {
    nixpkgs.overlays = [
      (final: prev: {
        cunicu = prev.callPackage ./cunicu.nix { };
      })
    ];

    users = {
      users.cunicu = {
        home = "/var/lib/cunicu";
        isSystemUser = true;
        group = "cunicu";
      };

      groups.cunicu = { };
    };

    environment.etc."cunicu.yaml" = {
      user = "cunicu";
      group = "cunicu";
      text = builtins.toJSON cfg.settings;
    };

    systemd = {
      services = {
        cunicu = {
          description = "WireGuard Interactive Connectivity Establishment";
          wants = [ "network-online.target" ];
          after = [ "network-online.target" ];
          wantedBy = [ "multi-user.target" ];
          serviceConfig = {
            Type = "simple";
						User = "cunicu";
						Group = "cunicu";
            ExecStart = "cunicu daemon";
          };
        };
      };
    };
  };
}
