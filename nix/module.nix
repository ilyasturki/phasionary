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

  # Passed as repeated flags rather than PHASIONARY_SERVE_ALLOWED_HOSTS: viper
  # reads a bound env var as one opaque string, so a comma-separated list
  # arrives as a single hostname and silently matches nothing.
  allowedHostFlags = lib.concatMapStringsSep " " (
    host: "--allowed-host ${lib.escapeShellArg host}"
  ) serveCfg.allowedHosts;
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
          `PHASIONARY_SERVE_TOKEN`. Required whenever `serve.enable` is set.

          It is not optional even on loopback: without it, `phasionary serve`
          generates a token on first run and prints it to stdout, which under
          systemd means the token is written to the journal and stays there.
        '';
      };

      allowedHosts = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        example = [ "phasionary.example.com" ];
        description = ''
          Additional `Host` header values to accept, passed as repeated
          `--allowed-host` flags.

          The server accepts IP literals and `localhost` unconditionally, so a
          LAN or Tailscale IP needs nothing here. Any other name — a reverse
          proxy's domain, a Tailscale MagicDNS name — is refused with 421 until
          it is listed, which is what closes off DNS rebinding.
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
          Whether to open `port` in the firewall. Ignored for loopback binds,
          where an open port would grant nothing — including behind a reverse
          proxy, which reaches the service without traversing the firewall.
        '';
      };
    };
  };

  config = lib.mkIf serveCfg.enable {
    assertions = [
      {
        assertion = serveCfg.tokenFile != null;
        message = ''
          services.phasionary.serve.tokenFile must be set.

          Without it phasionary generates a token on first run and prints it to
          stdout, which systemd captures into the journal — leaving a
          full-access credential readable by the journal-reading groups for the
          life of the log. This applies to loopback binds too, which is the
          usual shape behind a TLS reverse proxy.
        '';
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

    networking.firewall = lib.mkIf (serveCfg.openFirewall && !isLoopback) {
      allowedTCPPorts = [ serveCfg.port ];
    };

    # StateDirectory only covers the default path. A custom dataDir is named in
    # ReadWritePaths, and systemd refuses to start a unit whose ReadWritePaths
    # entry does not exist, so it has to be created here.
    systemd.tmpfiles.rules = lib.mkIf (serveCfg.dataDir != "/var/lib/phasionary") [
      "d ${serveCfg.dataDir} 0700 ${serveCfg.user} ${serveCfg.group} -"
    ];

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

        LoadCredential = [ "token:${toString serveCfg.tokenFile}" ];

        # Data files inherit this; config.json sets 0600 itself, but the store
        # is written with the process umask.
        UMask = "0077";

        NoNewPrivileges = true;
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ReadWritePaths = [ serveCfg.dataDir ];
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

      # tokenFile is asserted non-null above, so the credential always exists.
      script = ''
        export PHASIONARY_SERVE_TOKEN="$(cat "$CREDENTIALS_DIRECTORY/token")"
        exec ${cfg.package}/bin/phasionary serve ${allowedHostFlags}
      '';
    };
  };
}
