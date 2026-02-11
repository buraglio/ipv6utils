# IPv6 Utilities
This go project is the combination of some older python scripts that performed the same function, with a few more added features. Web version of this utlity can be found [here](https://tools.forwardingplane.net).

## Current capabilities: 
### IPv6 Subnet Generator & RFC 6052 Converter

A command-line utility for IPv6 subnet generation and IPv4/IPv6 address translation using RFC 6052.
The pre-compiled binary is compiled for Apple silicon.

## Features
- **IPv6 Subnet Generation**  
  - Generate subnets from an IPv6 prefix.  
  - Optionally limit output using `-l` flag.
- **IPv4 ↔ RFC 6052 IPv6 Conversion**  
  - Convert IPv4 to a synthesized IPv6 address (`-s IPv4`).  
  - Convert an RFC 6052 IPv6 address back to IPv4.  
- **Custom Prefix Support** (`-k`)  
  - Allows non-well-known prefixes.
- **Decode MAC address from SLAAC address** (-m)
  - Decode MAC addresses from non-privacy SLAAC addresses
- **Decode link local addresses** (-a)
  - Decode link local address to MAC, and MAC to link local
- **Generate ip6.arpa DNS names from IPv6 addresses** (`-ip6.arpa IPv6`)
  - Generate a partial reverse DNS name for use in an ip6.arpa zone file a specified prefix length (`-n`)
  - Generate a full reverse DNS name with `-n 0`
- **Display all IPv6 address formats** (`-f` / `-format`)
  - Expanded, compressed (RFC 5952), uppercase, URL bracket notation, dotted nibble, binary
  - Reverse DNS (ip6.arpa) and address type classification
  - All valid `::` compression permutations
---

## Installation

This should build on any system that can run golang, but has really only been tested on MacOS and Linux.

`git clone https://github.com/buraglio/ipv6utils.git`

`cd ipv6utils`

`go build -o ipv6utils ip6utils.go`

Move binary to wherever you want it to reside within your path, or reference it via a shell alias. 

## Homebrew Installation (MacOS)

`brew tap buraglio/ipv6utils`

`brew install ipv6utils`

## Use

Usage:
`./ipv6utils`
```
  -a string
        Alias for -local
  -f string
        Alias for -format
  -c    Alias for -count
  -count
        Display only the number of generated prefixes. (alias: -c)
  -format string
        Display all format representations of an IPv6 address. (alias: -f)
  -ip6.arpa string
        Generate a reverse ip6.arpa name for an IPv6 address. Uses -new-prefix-length as zone context.
  -k string
        Non-well-known prefix for RFC 6052 conversion. (default "64:ff9b::")
  -l int
        Limit the number of subnets displayed.
  -local string
        Link-local MAC or IPv6 to convert. (alias: -a)
  -m string
        SLAAC IPv6 address to decode MAC from.
  -n int
        Alias for -new-prefix-length (default 40)
  -new-prefix-length int
        New prefix length for subnet allocation. (alias: -n) (default 40)
  -o string
        Alias for -output
  -output string
        File to save the output subnets. (alias: -o)
  -p string
        Alias for -prefix (default "64:ff9b::")
  -prefix string
        IPv6 prefix for synthesis. (alias: -p) (default "64:ff9b::")
  -s string
        Source address for conversion.
```

### Link local decoder

`./ipv6utils.go -local fe80::0211:22ff:fe33:4455`

Output: 

`MAC from link-local: 00:11:22:33:44:55`

`./ipv6utils.go -local 00:11:22:33:44:55`

Output: 

`Link-local address: fe80::0211:22ff:fe33:4455`

### IPv4 → Synthesized IPv6

`./ipv6utils -s 8.8.8.8`

Output: 

`Converted IPv4 to synthesized IPv6: 64:ff9b::808:808`

### Synthesized IPv6 → IPv4

`./ipv6utils -s 64:ff9b::808:808`

Output:

`Converted synthesized IPv6 to IPv4: 8.8.8.8`

### Generate Subnets

`./ipv6utils -p 3fff::0::/32 -n 40 -l 5`

Output:

```
3fff::/40
3fff:0:100::/40
3fff:0:200::/40
3fff:0:300::/40
3fff:0:400::/40
```

### Save subnets to a File

`./ipv6utils -p 3fff:0::/32 -n 40 -o subnets.txt`

Output: 

`Subnets saved to subnets.txt`

### Decode MAC from SLAAC address (legacy)

Usage Example for -m:

`./ipv6utils -m 3fff:0::0200:5eff:fe00:5325`

Output: 

`Decoded MAC address: 00:00:5e:00:53:25`

### Generate ip6.arpa DNS names

Set prefix length to `0` for the full ip6.arpa name:

`./ipv6utils -ip6.arpa 2001:db8:abcd::0211:22ff:fe33:4455 -n 0`

Output:

`5.5.4.4.3.3.e.f.f.f.2.2.1.1.2.0.0.0.0.0.d.c.b.a.8.b.d.0.1.0.0.2.ip6.arpa.`

Set the prefix length of a DNS zone file for a partial name:

`./ipv6utils -ip6.arpa 2001:db8:abcd::0211:22ff:fe33:4455 -n 56`

Output:

`5.5.4.4.3.3.e.f.f.f.2.2.1.1.2.0.0.0`

### Display all IPv6 address formats

`./ipv6utils -f 2001:db8::1`

Output:

```
Expanded:       2001:0db8:0000:0000:0000:0000:0000:0001
Compressed:     2001:db8::1
Uppercase:      2001:DB8::1
URL format:     [2001:db8::1]
Dotted:         2.0.0.1.0.d.b.8.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1
Binary:         0010000000000001:0000110110111000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000001
Reverse DNS:    1.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa.
Address Type:   Documentation (2001:db8::/32)

Compression permutations:
  2001:db8::0:0:0:1
  2001:db8::0:0:1
  2001:db8::0:1
  2001:db8::1
  2001:db8:0::0:0:1
  2001:db8:0::0:1
  2001:db8:0::1
  2001:db8:0:0::0:1
  2001:db8:0:0::1
  2001:db8:0:0:0::1
```

With a prefix length:

`./ipv6utils -f 2001:db8::1/48`

Output appends `/48` to the Expanded and Compressed lines.
