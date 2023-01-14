# Installation

First, follow the [gokrazy installation instructions](https://gokrazy.org/quickstart/).

Then, add `github.com/gokrazy/syslogd/cmd/gokr-syslogd` to your gokrazy instance:

```
gok add github.com/gokrazy/syslogd/cmd/gokr-syslogd
```

Configure the listening address through a flag:

```json
{
    "Hostname": "router7",
    "Packages": [
        "github.com/gokrazy/fbstatus",
        "github.com/gokrazy/hello",
        "github.com/gokrazy/serial-busybox",
        "github.com/gokrazy/breakglass"
    ],
    "PackageConfig": {
        "github.com/gokrazy/syslogd/cmd/gokr-syslogd": {
            "CommandLineFlags": [
                "-listen=10.0.0.1:514"
            ]
        }
    },
}
```

## Usage Examples

To follow logs of a specific host live, install
https://github.com/gokrazy/breakglass for SSH access and use:

```shell
ssh router7 tail -f '/perm/syslogd/scan2drive/*.log'
```

You can also follow *all* logs:

```shell
ssh router7 tail -f '/perm/syslogd/*/*.log'
```

To search through old logs, grep through `*.log` for the last two days:

```shell
ssh router7 grep rror '/perm/syslogd/scan2drive/*.log'
```

…or `zstdgrep` through `*.log.gz` for older logs (not included in busybox
unfortunately):

```shell
sshfs router7:/perm/syslogd /mnt/syslogd
zstdgrep rror /mnt/syslogd/scan2drive/*.log.zst
```
