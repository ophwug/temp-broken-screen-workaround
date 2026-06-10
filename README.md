# Broken Screen openpilot Install and Enable Workaround

This is a temporary workaround tool for installing and enabling openpilot on comma devices where the setup screen may be malfunctioning.

Installing openpilot is only part of setup. openpilot still cannot be enabled until the on-device setup flow accepts terms and completes driver training. A broken or unusable screen can block those final setup gates. This tool works around that by installing over SSH and setting the same persistent params that make openpilot enable-able.

It scans your local network, finds comma/openpilot devices that accept SSH as `comma`, and installs a custom software URL without launching the on-device setup installer UI.

During install, the tool:

- downloads the same custom installer response that setup would use
- extracts the GitHub repo and branch from the installer
- applies the known tici/tizi/mici branch migration
- clones the selected branch into `/data/openpilot`
- attempts `git lfs pull` when Git LFS is available, but continues if it fails
- writes `/data/continue.sh`
- marks openpilot terms accepted and training completed using the installed branch's `system/version.py`, so openpilot can be enabled without finishing those screens on-device
- sets `SshEnabled=1`
- preserves `GithubSshKeys` from existing params, `/usr/comma/setup_keys`, or `/home/comma/.ssh/authorized_keys` when available
- restarts the comma wrapper

Use this only when you understand why bypassing setup/onboarding is needed.

The tool prints its own build version at startup. Include that line when sharing output.

It also writes a tee-style install log next to the executable, named like `install-20260610-130000.txt`, while still printing to the console.

## Windows Quick Start

1. Connect your Windows computer to the same Wi-Fi/network as the comma device.
2. If the device screen is usable enough to show the IP address, note it from **Advanced internet settings**.
3. Download the latest Windows executable:

   [temp-broken-screen-workaround.exe](https://github.com/ophwug/temp-broken-screen-workaround/releases/latest/download/temp-broken-screen-workaround.exe)

4. Double-click the executable.
5. Choose whether to enter the device IP address or scan the local network.
6. Enter the custom software URL, domain, or `owner/branch` when prompted.
7. Review the preflight output and let the install run.

If double-clicking closes too quickly, open Command Prompt in the download folder and run:

```cmd
temp-broken-screen-workaround.exe
```

If the install does not work, reflash the device at [flash.comma.ai](https://flash.comma.ai/) before retrying.

## Options

```text
--ip <addr>          Install to one known device IP instead of scanning.
--cidr <cidr>        Scan a specific IPv4 subnet, such as 192.168.1.0/24.
--parallel <n>       Maximum concurrent SSH probes. Default: 64.
--timeout <duration> SSH probe timeout. Default: 750ms.
--log <path>         Write the install log to a specific path. Use --log - to disable.
--custom-software-url <url>
                    Use this custom software URL instead of asking interactively.
```

Examples:

```cmd
temp-broken-screen-workaround.exe --ip 192.168.1.42 --custom-software-url openpilot-test.comma.ai
temp-broken-screen-workaround.exe --cidr 192.168.1.0/24
```

`--custom-software-url` accepts:

- a full URL, such as `https://openpilot-test.comma.ai`
- a domain, such as `openpilot-test.comma.ai`
- an owner/branch shorthand, such as `sunnypilot/dev`

For `owner/branch` shorthand, `installer.comma.ai` handles the repo selection. The usual fork-installer convention is that the repository name is hardcoded as `openpilot`; full custom URLs can still point at installers that encode a different GitHub repo.

## What SSH Persistence Does

openpilot uses the persistent `SshEnabled` param to control the developer SSH toggle, and `GithubSshKeys` as the authorized key material. This tool writes `SshEnabled=1` and preserves a key source into `GithubSshKeys` when one exists.

The preferred key source order is:

1. existing `/data/params/d/GithubSshKeys`
2. `/usr/comma/setup_keys`
3. `/home/comma/.ssh/authorized_keys`

If no key source is found, the tool still enables SSH but prints a warning that `GithubSshKeys` was not set.

## Recovery

If this workaround fails or leaves the device in a confusing state, reflash it with [flash.comma.ai](https://flash.comma.ai/) and retry from a clean setup state.

## macOS and Linux

The tool also builds for macOS and Linux:

```bash
curl -L https://github.com/ophwug/temp-broken-screen-workaround/releases/latest/download/temp-broken-screen-workaround-darwin -o temp-broken-screen-workaround
chmod +x temp-broken-screen-workaround
./temp-broken-screen-workaround
```

```bash
curl -L https://github.com/ophwug/temp-broken-screen-workaround/releases/latest/download/temp-broken-screen-workaround-linux -o temp-broken-screen-workaround
chmod +x temp-broken-screen-workaround
./temp-broken-screen-workaround
```

Automatic subnet detection is implemented for Windows, macOS, and Linux. If it cannot detect your active LAN, use `--ip` or `--cidr`.

## For Developers

Requirements:

- Go 1.22 or newer
- GitHub CLI if publishing releases manually

Commands:

```bash
go test ./...
make
make run
```

GitHub Actions builds the Windows, macOS, and Linux executables and publishes them to the `latest` release on every push to `main`.

## Credits

This was retooled from `ophwug/agnos-waiting-for-internet-debug`, which was retooled from `ophwug/c2-neos-alt-fix-install`, itself a Go evolution of the original NEOS manual installer by `jyoung8607`.
