# Installation

See https://gokrazy.org/quickstart/ first.

Then, add `github.com/gokrazy/syslogd/cmd/gokr-syslogd` to your `gokr-packer`
invocation.

Configure the listening address through a flag:
```shell
mkdir -p flags/github.com/gokrazy/syslogd/cmd/gokr-syslogd
echo '-listen=10.0.0.1:514' > flags/github.com/gokrazy/syslogd/cmd/gokr-syslogd/flags.txt
```
