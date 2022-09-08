# Installation

This guide shows how to install cunīcu.
cunīcu can be installed either from source, or from pre-built binary releases.

## From the Binary Releases

Every release of cunīcu provides binary releases for a variety of OSes.
These binary versions can be manually downloaded and installed.

## By Hand

1.  [Download your desired version](https://github.com/stv0g/cunicu/releases)
2.  Unzip it: `gunzip cunicu_0.0.1_linux_amd64.gz`
3.  Move the unzipped binary to its desired destination: `mv cunicu_0.0.1_linux_amd64 /usr/local/bin/cunicu`
4.  Make it executable: `chmod +x /usr/local/bin/cunicu`
5.  From there, you should be able to run the client and add the stable repo: `cunicu help`.

**Note:** cunīcu automated tests are performed for Linux, macOS and Windows on x86_64, ARMv6, ARMv8 amd ARM64 architectures.
Testing of other OSes are the responsibility of the community requesting cunīcu for the OS in question.

## From Script

cunīcu also has an installer script that will automatically grab the latest version of cunīcu and install it locally.

You can fetch that script, and then execute it locally.
It's well documented so that you can read through it and understand what it is doing before you run it.

```bash
curl -fsSL -o get_cunicu.sh https://raw.githubusercontent.com/stv0g/cunicu/master/scripts/get_cunicu.sh
chmod 700 get_cunicu.sh
./get_cunicu.sh
```

Yes, you can `curl https://raw.githubusercontent.com/stv0g/cunicu/master/scripts/get_cunicu.sh | bash` if you want to live on the edge.

## Through Package Managers

cunīcu provides the ability to install via operating system package managers.

### From Apt (Debian, Ubuntu)

```bash
sudo apt-get install apt-transport-https --yes
echo "deb [arch=$(dpkg --print-architecture) trusted=yes] https://packages.cunicu.li/apt/ /" | sudo tee /etc/apt/sources.list.d/cunicu.list
sudo apt-get update
sudo apt-get install cunicu
```

### From Yum (Redhat, Fedora, RockyLinux)

```bash
sudo cat > /etc/yum.repos.d/cunicu.repo <<EOF
[cunicu]
name=cunicu
baseurl=https://packages.cunicu/yum/
enabled=1
gpgcheck=0
EOF
sudo yum install cunīcu
```

## From Source (all)

Building cunīcu is fairly easy and allows you to install the latest unreleased version.

You must have a working Go environment.

```bash
go install github.com/stv0g/cunicu/cmd/cunicu@latest
```

If required, it will fetch the dependencies and cache them, and validate configuration.
It will then compile cunīcu and place it in `${GOPATH}/bin/cunicu`.

## Conclusion

In most cases, installation is as simple as getting a pre-built cunīcu binary.
This document covers additional cases for those who want to do more sophisticated things with cunīcu.

Once you have cunīcu successfully installed, you can move on to [using cunīcu](Usage.md) to setup your mesh VPN network.