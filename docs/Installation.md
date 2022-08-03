# Installation

This guide shows how to install ɯice.
ɯice can be installed either from source, or from pre-built binary releases.

## From the Binary Releases

Every release of ɯice provides binary releases for a variety of OSes.
These binary versions can be manually downloaded and installed.

## By Hand

1. [Download your desired version](https://github.com/stv0g/wice/releases)
2. Unzip it: `gunzip wice_0.0.1_linux_amd64.gz`
3. Move the unzipped binary to its desired destination: `mv wice_0.0.1_linux_amd64 /usr/local/bin/wice`
5. Make it executable: `chmod +x /usr/local/bin/wice`
6. From there, you should be able to run the client and add the stable repo: `wice help`.

**Note:** ɯice automated tests are performed for Linux, macOS and Windows on x86_64, ARMv6, ARMv8 amd ARM64 architectures.
Testing of other OSes are the responsibility of the community requesting ɯice for the OS in question.

## From Script

ɯice also has an installer script that will automatically grab the latest version of ɯice and install it locally.

You can fetch that script, and then execute it locally.
It's well documented so that you can read through it and understand what it is doing before you run it.

```bash
$ curl -fsSL -o get_wice.sh https://raw.githubusercontent.com/stv0g/wice/master/scripts/get_wice.sh
$ chmod 700 get_wice.sh
$ ./get_wice.sh
```

Yes, you can `curl https://raw.githubusercontent.com/stv0g/wice/master/scripts/get_wice.sh | bash` if you want to live on the edge.

## Through Package Managers

ɯice provides the ability to install via operating system package managers.

### From Apt (Debian, Ubuntu)

```bash
sudo apt-get install apt-transport-https --yes
echo "deb [arch=$(dpkg --print-architecture) trusted=yes] https://packages.riasc.eu/apt/ /" | sudo tee /etc/apt/sources.list.d/riasc.list
sudo apt-get update
sudo apt-get install wice
```

### From Yum (Redhat, Fedora, RockyLinux)

```bash
sudo cat > /etc/yum.repos.d/riasc.repo <<EOF
[riasc]
name=RIasC
baseurl=https://packages.riasc.eu/yum/
enabled=1
gpgcheck=0
EOF
sudo yum install wice
```

## From Source (all)

Building ɯice is fairly easy and allows you to install the latest unreleased version.

You must have a working Go environment.

```
$ go install riasc.eu/wice/cmd@latest
```

If required, it will fetch the dependencies and cache them, and validate configuration.
It will then compile ɯice and place it in `${GOPATH}/bin/wice`.

##  Conclusion

In most cases, installation is as simple as getting a pre-built ɯice binary.
This document covers additional cases for those who want to do more sophisticated things with ɯice.

Once you have ɯice successfully installed, you can move on to [using ɯice](Usage.md) to setup your mesh VPN network.