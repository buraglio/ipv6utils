# IPv6 Utilities
This go project is the combination of two python scripts that performed the same function.

## Current capabilities: 
### IPv6 Subnet Generator & RFC 6052 Converter

A command-line utility for IPv6 subnet generation and IPv4/IPv6 address translation using RFC 6052.
The pre-compiled binary is compiled for Apple silicon.

## 🚀 Features
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
---

## Installation

`go build -o ipv6utils ip6utils.go`

## Use

Usage:
`./ipv6utils`
```
  -a string
        Alias for -local
  -c    Alias for -count
  -count
        Display only the number of generated prefixes. (alias: -c)
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

`./ipv6utils -p 3ffe::0::/32 -n 40 -l 5`

Output:

```
3ffe::/40
3ffe:0:100::/40
3ffe:0:200::/40
3ffe:0:300::/40
3ffe:0:400::/40
```

### Save subnets to a File

`./ipv6utils -p 3ffe:0::/32 -n 40 -o subnets.txt`

Output: 

`Subnets saved to subnets.txt`

### Decode MAC from SLAAC address (legacy)

Usage Example for -m:

`./ipv6utils -m 3ffe:0::0200:5eff:fe00:5325`

Output: 

`Decoded MAC address: 00:00:5e:00:53:25`

