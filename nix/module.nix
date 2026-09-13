{ self }:
{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.phasionary-server;
  defaultPackage = self.packages.${pkgs.stdenv.hostPlatform.system}.phasionary-server;

  isLoopback = cfg.host == "127.0.0.1" || cfg.host == "::1" || cfg.host == "localhost";
in
{
  options.services.phasionary-server = {
    enable = lib.mkEnableOption "phasionary sync server";

    package = lib.mkOption {
      type = lib.types.package;
      default = defaultPackage;
      defaultText = lib.literalExpression "phasionary.packages.\${system}.phasionary-server";
      description = "The phasionary-server package to use.";
    };

    host = lib.mkOption {
      type = lib.types.str;
      default = "127.0.0.1";
      example = "100.64.0.1";
      description = ''
        Host/IP the server binds to. The server speaks plain HTTP and every
        request carries a device's bearer token, so bind to a private
        network address (Tailscale, WireGuard) or keep loopback and put a
        TLS reverse proxy in front.
      '';
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 7777;
      description = "TCP port the server listens on.";
    };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/phasionary-server";
      description = ''
        Directory holding the SQLite database. Enrollment codes are minted
        against the same database, so enroll a device with
        `sudo -u ''${user} phasionary-server --data ''${dataDir} enroll`.
      '';
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "phasionary";
      description = "User account under which the server runs.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "phasionary";
      description = "Group under which the server runs.";
    };

    allowedHosts = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [ ];
      example = [ "phas.example.net" ];
      description = ''
        Hostnames a browser may reach the web app by; IP literals and
        `localhost` always pass. A named host has to be listed (a DNS-rebinding
        guard): set it to the name your reverse proxy serves.
      '';
    };

    openFirewall = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = ''
        Whether to open `port` in the firewall. Ignored for loopback binds,
        where an open port would grant nothing — including behind a reverse
        proxy, which reaches the service without traversing the firewall.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];

    users.users = lib.mkIf (cfg.user == "phasionary") {
      phasionary = {
        isSystemUser = true;
        group = cfg.group;
        home = cfg.dataDir;
      };
    };

    users.groups = lib.mkIf (cfg.group == "phasionary") {
      phasionary = { };
    };

    networking.firewall = lib.mkIf (cfg.openFirewall && !isLoopback) {
      allowedTCPPorts = [ cfg.port ];
    };

    # systemd refuses a unit whose ReadWritePaths entry is missing; StateDirectory
    # only creates the default path.
    systemd.tmpfiles.rules = lib.mkIf (cfg.dataDir != "/var/lib/phasionary-server") [
      "d ${cfg.dataDir} 0700 ${cfg.user} ${cfg.group} -"
    ];

    systemd.services.phasionary-server = {
      description = "Phasionary sync server";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      environment = {
        PHASIONARY_SERVER_DATA_PATH = cfg.dataDir;
        PHASIONARY_SERVER_HOST = cfg.host;
        PHASIONARY_SERVER_PORT = toString cfg.port;
        PHASIONARY_SERVER_ALLOWED_HOSTS = lib.concatStringsSep "," cfg.allowedHosts;
      };

      serviceConfig = {
        ExecStart = "${cfg.package}/bin/phasionary-server";
        User = cfg.user;
        Group = cfg.group;
        Restart = "on-failure";
        RestartSec = "5s";

        StateDirectory = lib.mkIf (cfg.dataDir == "/var/lib/phasionary-server") "phasionary-server";

        UMask = "0077";

        NoNewPrivileges = true;
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadWritePaths = [ cfg.dataDir ];
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectKernelLogs = true;
        ProtectControlGroups = true;
        ProtectClock = true;
        ProtectHostname = true;
        ProtectProc = "invisible";
        ProcSubset = "pid";
        RestrictAddressFamilies = [
          "AF_UNIX"
          "AF_INET"
          "AF_INET6"
        ];
        RestrictNamespaces = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        LockPersonality = true;
        CapabilityBoundingSet = [ "" ];
        SystemCallArchitectures = "native";
        SystemCallFilter = [
          "@system-service"
          "~@privileged"
          "~@resources"
        ];
      };
    };
  };
}
