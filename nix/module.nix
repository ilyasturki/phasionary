{ self }:
{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.phasionary;
  serveCfg = cfg.serve;
  defaultPackage = self.packages.${pkgs.stdenv.hostPlatform.system}.phasionary;

  parsedAddr =
    let
      bracketed = builtins.match "[[](.+)[]]:([0-9]+)" serveCfg.addr;
      parts = lib.splitString ":" serveCfg.addr;
    in
    if bracketed != null then {
      host = lib.elemAt bracketed 0;
      port = lib.toInt (lib.elemAt bracketed 1);
    } else {
      host = lib.head parts;
      port = lib.toInt (lib.last parts);
    };

  isLoopback =
    parsedAddr.host == "127.0.0.1"
    || parsedAddr.host == "::1"
    || parsedAddr.host == "localhost";
in
{
  options.services.phasionary = {
    package = lib.mkOption {
      type = lib.types.package;
      default = defaultPackage;
      defaultText = lib.literalExpression "phasionary.packages.\${system}.phasionary";
      description = "The phasionary package to use.";
    };

    serve = {
      enable = lib.mkEnableOption "Phasionary htmx web server";

      addr = lib.mkOption {
        type = lib.types.str;
        default = "127.0.0.1:7777";
        example = "0.0.0.0:7777";
        description = ''
          Listen address (host:port) passed to `phasionary serve --addr`.
          A token is required for any non-loopback bind.
        '';
      };

      tokenFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        example = "/run/secrets/phasionary-token";
        description = ''
          Path to a file containing the auth token. The file is read by systemd
          via `LoadCredential` and exposed to the service as
          `PHASIONARY_SERVE_TOKEN`. Required when `addr` is not loopback.
        '';
      };

      dataDir = lib.mkOption {
        type = lib.types.path;
        default = "/var/lib/phasionary";
        description = ''
          Directory where Phasionary stores project data. Exported as
          `PHASIONARY_DATA_PATH`.
        '';
      };

      user = lib.mkOption {
        type = lib.types.str;
        default = "phasionary";
        description = "User account under which phasionary runs.";
      };

      group = lib.mkOption {
        type = lib.types.str;
        default = "phasionary";
        description = "Group under which phasionary runs.";
      };

      openFirewall = lib.mkOption {
        type = lib.types.bool;
        default = false;
        description = ''
          Whether to open the TCP port in `addr` in the firewall. Only takes
          effect for non-loopback binds.
        '';
      };
    };
  };

  config = lib.mkIf serveCfg.enable {
    assertions = [
      {
        assertion = isLoopback || serveCfg.tokenFile != null;
        message = "services.phasionary.serve.tokenFile must be set when addr is not loopback.";
      }
    ];

    users.users = lib.mkIf (serveCfg.user == "phasionary") {
      phasionary = {
        isSystemUser = true;
        group = serveCfg.group;
        home = serveCfg.dataDir;
      };
    };

    users.groups = lib.mkIf (serveCfg.group == "phasionary") {
      phasionary = { };
    };

    networking.firewall = lib.mkIf serveCfg.openFirewall {
      allowedTCPPorts = [ parsedAddr.port ];
    };

    systemd.services.phasionary-serve = {
      description = "Phasionary htmx web server";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      environment = {
        PHASIONARY_DATA_PATH = serveCfg.dataDir;
        PHASIONARY_SERVE_ADDR = serveCfg.addr;
      };

      serviceConfig = {
        User = serveCfg.user;
        Group = serveCfg.group;
        Restart = "on-failure";
        RestartSec = "5s";

        StateDirectory = lib.mkIf (serveCfg.dataDir == "/var/lib/phasionary") "phasionary";

        LoadCredential = lib.mkIf (serveCfg.tokenFile != null) [
          "token:${toString serveCfg.tokenFile}"
        ];

        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadWritePaths = [ serveCfg.dataDir ];
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [
          "AF_UNIX"
          "AF_INET"
          "AF_INET6"
        ];
        RestrictNamespaces = true;
        LockPersonality = true;
        SystemCallArchitectures = "native";
      };

      script =
        if serveCfg.tokenFile != null then
          ''
            export PHASIONARY_SERVE_TOKEN="$(cat "$CREDENTIALS_DIRECTORY/token")"
            exec ${cfg.package}/bin/phasionary serve
          ''
        else
          ''
            exec ${cfg.package}/bin/phasionary serve
          '';
    };
  };
}
