# IPv6 Utilities

A Go CLI for IPv6 address analysis, subnet generation, and address translation. Combines and extends several older Python scripts with additional features. A web version of this utility is available at [tools.forwardingplane.net](https://tools.forwardingplane.net).

---

## Features

- **IPv6 Subnet Generation** — generate subnets from an IPv6 prefix with optional limit and file output
- **IPv4 ↔ RFC 6052 IPv6 Conversion** — synthesize and extract IPv4 addresses using NAT64 prefixes
- **Custom Prefix Support** — non-well-known RFC 6052 prefixes via `-k`
- **SLAAC MAC Decode** — extract the original MAC from a non-privacy EUI-64 SLAAC address
- **Link-Local ↔ MAC Conversion** — bidirectional EUI-64 link-local address conversion
- **Reverse DNS Generation** — full or partial `ip6.arpa` names for zone files
- **Address Format Display** — all representations of an IPv6 address in one shot:
  - Expanded, compressed (RFC 5952), uppercase, URL bracket, dotted nibble, binary
  - Reverse DNS name and address type classification
  - IPv4-in-IPv6 mixed notation for IPv4-mapped addresses (`::ffff:x.x.x.x`)
  - When a prefix length is supplied: **network address, host ID, and network range**
  - All valid `::` compression permutations

---

## Installation

### Build from source

Requires Go 1.22+. Builds on any platform Go supports.

```sh
git clone https://github.com/buraglio/ipv6utils.git
cd ipv6utils
go build -o ipv6utils ipv6utils.go
```

Move the binary wherever you need it, or reference it via a shell alias.

### Makefile targets

```sh
make all               # native binary with version embedded
make dist              # all supported platforms → ./dist/
make linux/arm64       # single platform build
make clean             # remove ./dist/ and local binary
make help              # list targets
```

Override the version string at build time:

```sh
make all VERSION=5
```

Cross-compiled binaries are written to `./dist/` as `ipv6utils_OS_ARCH[.exe]`.
Supported platforms include Linux (x86, ARM, MIPS, RISC-V, s390x, ppc64le),
macOS (Intel + Apple Silicon), Windows, FreeBSD, OpenBSD, NetBSD, DragonFly BSD,
and Solaris/illumos.

### Homebrew (macOS)

```sh
brew tap buraglio/ipv6utils
brew install ipv6utils
```

---

## Usage

```sh
./ipv6utils [OPTIONS]
```

| Flag | Alias | Description |
| --- | --- | --- |
| `-format ADDR[/N]` | `-f` | Display all format representations of an IPv6 address. Supply a prefix length to also show network range and host ID. |
| `-s ADDR` | | Convert IPv4↔IPv6 (direction auto-detected). Uses `-k` prefix. |
| `-m ADDR` | | Decode MAC address from a SLAAC (EUI-64) IPv6 address. |
| `-local ADDR` | `-a` | Convert link-local ↔ MAC (direction auto-detected). |
| `-ip6.arpa ADDR` | | Generate a reverse DNS name. Use `-n` for zone context. |
| `-prefix PREFIX` | `-p` | Base IPv6 prefix for subnet generation. (default: `64:ff9b::`) |
| `-new-prefix-length N` | `-n` | New prefix length for subnets or ip6.arpa zone context. (default: `40`) |
| `-limit N` | `-l` | Limit subnet output to N entries. |
| `-count` | `-c` | Print only the count of subnets that would be generated. |
| `-output FILE` | `-o` | Save generated subnets to a file. |
| `-k PREFIX` | | Non-well-known RFC 6052 prefix for synthesis. (default: `64:ff9b::`) |
| `-version` | `-v` | Print version and exit. |

---

## Examples

### Display all address formats

Without a prefix length — shows representations only:

```sh
./ipv6utils -f 2001:db8::1
```

```text
Expanded:       2001:0db8:0000:0000:0000:0000:0000:0001
Compressed:     2001:db8::1
Uppercase:      2001:DB8::1
URL format:     [2001:db8::1]
Dotted:         2.0.0.1.0.d.b.8.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1
Binary:         0010000000000001:0000110110111000:...
Reverse DNS:    1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.
Address Type:   Documentation (2001:db8::/32)

Compression permutations:
  2001:db8::1
  2001:db8:0::1
  ...
```

With a prefix length — also shows network address, host ID, and network range:

```sh
./ipv6utils -f 2001:db8::1/48
```

```text
Expanded:       2001:0db8:0000:0000:0000:0000:0000:0001/48
Compressed:     2001:db8::1/48
Uppercase:      2001:DB8::1
URL format:     [2001:db8::1]
Dotted:         2.0.0.1.0.d.b.8...
Binary:         0010000000000001:0000110110111000:...
Reverse DNS:    1.0.0.0...ip6.arpa.
Address Type:   Documentation (2001:db8::/32)

Network:        2001:0db8:0000:0000:0000:0000:0000:0000/48
Host ID:        ::1/48
Network range:  2001:0db8:0000:0000:0000:0000:0000:0000 -
                2001:0db8:0000:ffff:ffff:ffff:ffff:ffff

Compression permutations:
  2001:db8::1
  ...
```

IPv4-mapped addresses include a mixed-notation line:

```sh
./ipv6utils -f ::ffff:192.0.2.1
```

```text
Expanded:       0000:0000:0000:0000:0000:ffff:c000:0201
Compressed:     192.0.2.1
...
Address Type:   IPv4-Mapped (::ffff:0:0/96)
IPv4-in-IPv6:   ::ffff:192.0.2.1
```

### IPv4 → Synthesized IPv6

```sh
./ipv6utils -s 8.8.8.8
```

```text
Converted IPv4 to synthesized IPv6: 64:ff9b::808:808
```

### Synthesized IPv6 → IPv4

```sh
./ipv6utils -s 64:ff9b::808:808
```

```text
Converted synthesized IPv6 to IPv4: 8.8.8.8
```

### Link-local ↔ MAC

```sh
./ipv6utils -local 00:11:22:33:44:55
```

```text
Link-local address: fe80::0211:22ff:fe33:4455
```

```sh
./ipv6utils -local fe80::0211:22ff:fe33:4455
```

```text
MAC from link-local: 00:11:22:33:44:55
```

### Decode MAC from SLAAC address

```sh
./ipv6utils -m 3fff:0::0200:5eff:fe00:5325
```

```text
Decoded MAC address: 00:00:5e:00:53:25
```

### Generate subnets

```sh
./ipv6utils -p 3fff::/32 -n 40 -l 5
```

```text
Generating 256 prefixes...
3fff::/40
3fff:0:100::/40
3fff:0:200::/40
3fff:0:300::/40
3fff:0:400::/40
```

Count only:

```sh
./ipv6utils -p 3fff::/32 -n 40 -c
```

```text
Number of prefixes: 256
```

Save to file:

```sh
./ipv6utils -p 3fff::/32 -n 40 -o subnets.txt
```

### Reverse DNS names

Full `ip6.arpa` name (`-n 0`):

```sh
./ipv6utils -ip6.arpa 2001:db8:abcd::0211:22ff:fe33:4455 -n 0
```

```text
5.5.4.4.3.3.e.f.f.f.2.2.1.1.2.0.0.0.0.0.d.c.b.a.8.b.d.0.1.0.0.2.ip6.arpa.
```

Partial name for a `/56` zone file:

```sh
./ipv6utils -ip6.arpa 2001:db8:abcd::0211:22ff:fe33:4455 -n 56
```

```text
5.5.4.4.3.3.e.f.f.f.2.2.1.1.2.0.0.0
```

### Version

```sh
./ipv6utils -v
```

```text
ipv6utils 4
```
