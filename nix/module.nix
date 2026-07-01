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

  isLoopback =
    serveCfg.host == "127.0.0.1"
    || serveCfg.host == "::1"
    || serveCfg.host == "localhost";
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
      enable = lib.mkEnableOption "Phasionary JSON API server";

      host = lib.mkOption {
        type = lib.types.str;
        default = "127.0.0.1";
        example = "0.0.0.0";
        description = ''
          Host/IP the server binds to, passed to `phasionary serve --host`.
          A token is required for any non-loopback bind.
        '';
      };

      port = lib.mkOption {
        type = lib.types.port;
        default = 7777;
        description = ''
          TCP port the server listens on, passed to `phasionary serve --port`.
        '';
      };

      tokenFile = lib.mkOption {
        type = lib.types.nullOr lib.types.path;
        default = null;
        example = "/run/secrets/phasionary-token";
        description = ''
          Path to a file containing the auth token. The file is read by systemd
          via `LoadCredential` and exposed to the service as
          `PHASIONARY_SERVE_TOKEN`. Required when `host` is not loopback.
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
          Whether to open `port` in the firewall. Only takes effect for
          non-loopback binds.
        '';
      };
    };
  };

  config = lib.mkIf serveCfg.enable {
    assertions = [
      {
        assertion = isLoopback || serveCfg.tokenFile != null;
        message = "services.phasionary.serve.tokenFile must be set when host is not loopback.";
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
      allowedTCPPorts = [ serveCfg.port ];
    };

    systemd.services.phasionary-serve = {
      description = "Phasionary JSON API server";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      environment = {
        PHASIONARY_DATA_PATH = serveCfg.dataDir;
        PHASIONARY_SERVE_HOST = serveCfg.host;
        PHASIONARY_SERVE_PORT = toString serveCfg.port;
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
