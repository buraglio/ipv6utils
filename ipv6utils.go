// SPDX-License-Identifier: BSD-3-Clause-LBNL
// Copyright (C) buraglio@forwardingplane.net

package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"slices"
	"sort"
	"strings"
)

// isNibbleAligned checks whether the prefix length is on a nibble boundary (multiple of 4).
func isNibbleAligned(prefixLength int) bool {
	return prefixLength%4 == 0
}

// countSubnets calculates how many subnets would be generated from the original prefix to the new length.
func countSubnets(prefix string, newPrefixLength int) (int, error) {
	_, ipnet, err := net.ParseCIDR(prefix)
	if err != nil {
		return 0, fmt.Errorf("invalid prefix: %v", err)
	}
	currentPrefixLength, _ := ipnet.Mask.Size()
	if newPrefixLength <= currentPrefixLength {
		return 0, fmt.Errorf("new prefix length must be larger than the current prefix length")
	}
	return 1 << (newPrefixLength - currentPrefixLength), nil
}

// generateSubnets produces subnets of a specified length from a base prefix with optional output limiting.
func generateSubnets(prefix string, newPrefixLength int, limit int) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(prefix)
	if err != nil {
		return nil, fmt.Errorf("invalid prefix: %v", err)
	}
	if !isNibbleAligned(newPrefixLength) {
		log.Println("Warning: new prefix length is not on a nibble boundary")
	}
	currentPrefixLength, _ := ipnet.Mask.Size()
	if newPrefixLength <= currentPrefixLength {
		return nil, fmt.Errorf("new prefix length must be larger than the current prefix length")
	}
	subnetCount := 1 << (newPrefixLength - currentPrefixLength)
	fmt.Printf("Generating %d prefixes...\n", subnetCount)
	subnets := []string{}
	prefixIP := ipnet.IP.Mask(ipnet.Mask)
	increment := big.NewInt(1)
	increment.Lsh(increment, uint(128-newPrefixLength))
	for i := 0; i < subnetCount; i++ {
		subnets = append(subnets, fmt.Sprintf("%s/%d", prefixIP, newPrefixLength))
		prefixIP = addBigIntToIP(prefixIP, increment)
		if limit > 0 && len(subnets) >= limit {
			break
		}
	}
	sort.Slice(subnets, func(i, j int) bool {
		ip1 := net.ParseIP(strings.Split(subnets[i], "/")[0])
		ip2 := net.ParseIP(strings.Split(subnets[j], "/")[0])
		return bytes.Compare(ip1, ip2) < 0
	})
	return subnets, nil
}

// addBigIntToIP adds a big integer to an IPv6 address and returns the resulting IP.
func addBigIntToIP(ip net.IP, value *big.Int) net.IP {
	ipInt := big.NewInt(0)
	ipInt.SetBytes(ip.To16())
	ipInt.Add(ipInt, value)
	newIP := ipInt.Bytes()
	if len(newIP) < net.IPv6len {
		padding := make([]byte, net.IPv6len-len(newIP))
		newIP = append(padding, newIP...)
	}
	return newIP
}

// synthesizedToIPv4 converts an RFC 6052 synthesized IPv6 address to its embedded IPv4 address.
func synthesizedToIPv4(synthesizedAddr string) (string, error) {
	ip := net.ParseIP(synthesizedAddr)
	if ip == nil || ip.To16() == nil {
		return "", fmt.Errorf("invalid RFC 6052 synthesized address")
	}
	ipv4 := ip[12:16]
	if len(ipv4) != 4 {
		return "", fmt.Errorf("not a valid synthesized IPv6 address containing IPv4")
	}
	return fmt.Sprintf("%d.%d.%d.%d", ipv4[0], ipv4[1], ipv4[2], ipv4[3]), nil
}

// ipv4ToSynthesized converts an IPv4 address into an RFC 6052 synthesized IPv6 address using the provided prefix.
func ipv4ToSynthesized(ipv4Addr string, prefix string) (string, error) {
	ip := net.ParseIP(ipv4Addr)
	if ip == nil || ip.To4() == nil {
		return "", fmt.Errorf("invalid IPv4 address")
	}
	prefixIP := net.ParseIP(prefix)
	if prefixIP == nil {
		return "", fmt.Errorf("invalid IPv6 prefix")
	}
	ipv6Addr := make(net.IP, net.IPv6len)
	copy(ipv6Addr, prefixIP.To16())
	copy(ipv6Addr[12:], ip.To4())
	return ipv6Addr.String(), nil
}

// decodeMACFromSLAAC extracts a MAC address from a given SLAAC IPv6 address.
func decodeMACFromSLAAC(ipv6 string) (string, error) {
	ip := net.ParseIP(ipv6)
	if ip == nil || ip.To16() == nil {
		return "", fmt.Errorf("invalid SLAAC IPv6 address")
	}
	interfaceID := ip[8:]
	if len(interfaceID) < 8 || interfaceID[3] != 0xFF || interfaceID[4] != 0xFE {
		return "", fmt.Errorf("not a valid EUI-64 SLAAC address (missing FFFE)")
	}
	mac := []byte{
		interfaceID[0] ^ 0x02,
		interfaceID[1],
		interfaceID[2],
		interfaceID[5],
		interfaceID[6],
		interfaceID[7],
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]), nil
}

// macToLinkLocal converts a MAC address into an EUI-64 formatted link-local IPv6 address.
func macToLinkLocal(mac string) (string, error) {
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return "", fmt.Errorf("invalid MAC address format")
	}
	var b [6]byte
	for i := 0; i < 6; i++ {
		_, err := fmt.Sscanf(parts[i], "%x", &b[i])
		if err != nil {
			return "", err
		}
	}
	b[0] ^= 0x02
	return fmt.Sprintf("fe80::%02x%02x:%02xff:fe%02x:%02x%02x", b[0], b[1], b[2], b[3], b[4], b[5]), nil
}

// linkLocalToMAC extracts a MAC address from a link-local EUI-64 formatted IPv6 address.
func linkLocalToMAC(ipv6 string) (string, error) {
	ip := net.ParseIP(ipv6)
	if ip == nil || ip.To16() == nil || !strings.HasPrefix(ipv6, "fe80") {
		return "", fmt.Errorf("invalid IPv6 address")
	}
	interfaceID := ip[8:]
	if len(interfaceID) < 8 || interfaceID[3] != 0xFF || interfaceID[4] != 0xFE {
		return "", fmt.Errorf("not a valid EUI-64 link-local address")
	}
	mac := []byte{
		interfaceID[0] ^ 0x02,
		interfaceID[1],
		interfaceID[2],
		interfaceID[5],
		interfaceID[6],
		interfaceID[7],
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]), nil
}

// ipv6ToArpa returns a dot-separated, reversed string of nibbles for use in constructing
// a reverse ip6.arpa DNS record. Assumes a zone context matching the prefix and outputs
// only the non-prefix nibbles. If the prefix is 0, the full ip6.arpa name is returned.
func ipv6ToArpa(ipv6 string, prefixLength int) (string, error) {
	ip := net.ParseIP(ipv6)
	if ip == nil || ip.To16() == nil {
		return "", fmt.Errorf("Invalid IP address: %s", ipv6)
	}
	if prefixLength < 0 || prefixLength > 128 {
		return "", fmt.Errorf("Invalid prefix length: %d", prefixLength)
	}
	if !isNibbleAligned(prefixLength) {
		log.Println("Warning: prefix length is not on a nibble boundary")
	}
	nibbles := strings.Split(hex.EncodeToString(ip.To16()), "")
	slices.Reverse(nibbles)
	trim := 32 - (prefixLength / 4)
	if trim >= 0 && trim < 32 {
		return strings.Join(nibbles[:trim], "."), nil
	}
	return fmt.Sprintf("%s.ip6.arpa.", strings.Join(nibbles, ".")), nil
}

// parseIPv6WithOptionalPrefix splits an input string into an IPv6 address and optional prefix length.
// Returns prefix length of -1 when no prefix is provided.
func parseIPv6WithOptionalPrefix(input string) (net.IP, int, error) {
	if input == "" {
		return nil, -1, fmt.Errorf("empty input")
	}

	addr := input
	prefixLen := -1

	if idx := strings.LastIndex(input, "/"); idx != -1 {
		addr = input[:idx]
		var pfx int
		n, err := fmt.Sscanf(input[idx+1:], "%d", &pfx)
		if err != nil || n != 1 {
			return nil, -1, fmt.Errorf("invalid prefix length: %s", input[idx+1:])
		}
		if pfx < 0 || pfx > 128 {
			return nil, -1, fmt.Errorf("prefix length must be between 0 and 128, got %d", pfx)
		}
		prefixLen = pfx
	}

	if !strings.Contains(addr, ":") {
		return nil, -1, fmt.Errorf("not an IPv6 address: %s", addr)
	}

	ip := net.ParseIP(addr)
	if ip == nil {
		return nil, -1, fmt.Errorf("invalid IPv6 address: %s", addr)
	}
	ip = ip.To16()
	if ip == nil {
		return nil, -1, fmt.Errorf("invalid IPv6 address: %s", addr)
	}

	return ip, prefixLen, nil
}

// expandIPv6 returns the fully expanded 8-group, zero-padded IPv6 address.
func expandIPv6(ip net.IP) string {
	b := ip.To16()
	return fmt.Sprintf("%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x:%02x%02x",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7],
		b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15])
}

// compressIPv6 returns the RFC 5952 canonical compressed form.
func compressIPv6(ip net.IP) string {
	return ip.String()
}

// uppercaseIPv6 returns the compressed form in uppercase.
func uppercaseIPv6(ip net.IP) string {
	return strings.ToUpper(ip.String())
}

// urlIPv6 returns the address in URL bracket notation.
func urlIPv6(ip net.IP) string {
	return fmt.Sprintf("[%s]", ip.String())
}

// dottedIPv6 returns each nibble of the address separated by dots.
func dottedIPv6(ip net.IP) string {
	nibbles := strings.Split(hex.EncodeToString(ip.To16()), "")
	return strings.Join(nibbles, ".")
}

// binaryIPv6 returns the address as 8 colon-separated 16-bit binary groups.
func binaryIPv6(ip net.IP) string {
	b := ip.To16()
	groups := make([]string, 8)
	for i := 0; i < 8; i++ {
		val := uint16(b[i*2])<<8 | uint16(b[i*2+1])
		groups[i] = fmt.Sprintf("%016b", val)
	}
	return strings.Join(groups, ":")
}

// classifyIPv6 returns a human-readable string describing the address type.
func classifyIPv6(ip net.IP) string {
	b := ip.To16()

	// Unspecified (::)
	if ip.Equal(net.IPv6unspecified) {
		return "Unspecified (::)"
	}
	// Loopback (::1)
	if ip.Equal(net.IPv6loopback) {
		return "Loopback (::1)"
	}
	// IPv4-Mapped (::ffff:0:0/96)
	allZero := true
	for i := 0; i < 10; i++ {
		if b[i] != 0 {
			allZero = false
			break
		}
	}
	if allZero && b[10] == 0xff && b[11] == 0xff {
		return "IPv4-Mapped (::ffff:0:0/96)"
	}
	// NAT64 Well-Known Prefix (64:ff9b::/96)
	if b[0] == 0x00 && b[1] == 0x64 && b[2] == 0xff && b[3] == 0x9b &&
		b[4] == 0x00 && b[5] == 0x00 && b[6] == 0x00 && b[7] == 0x00 &&
		b[8] == 0x00 && b[9] == 0x00 && b[10] == 0x00 && b[11] == 0x00 {
		return "NAT64 Well-Known Prefix (64:ff9b::/96)"
	}
	// NAT64 Network-Specific (64:ff9b:1::/48)
	if b[0] == 0x00 && b[1] == 0x64 && b[2] == 0xff && b[3] == 0x9b &&
		b[4] == 0x00 && b[5] == 0x01 {
		return "NAT64 Network-Specific (64:ff9b:1::/48)"
	}
	// Discard-Only (100::/64)
	if b[0] == 0x01 && b[1] == 0x00 &&
		b[2] == 0x00 && b[3] == 0x00 && b[4] == 0x00 && b[5] == 0x00 &&
		b[6] == 0x00 && b[7] == 0x00 {
		return "Discard-Only (100::/64)"
	}
	// Teredo (2001:0000::/32)
	if b[0] == 0x20 && b[1] == 0x01 && b[2] == 0x00 && b[3] == 0x00 {
		return "Teredo (2001:0000::/32)"
	}
	// Documentation (2001:db8::/32)
	if b[0] == 0x20 && b[1] == 0x01 && b[2] == 0x0d && b[3] == 0xb8 {
		return "Documentation (2001:db8::/32)"
	}
	// 6to4 (2002::/16)
	if b[0] == 0x20 && b[1] == 0x02 {
		return "6to4 (2002::/16)"
	}
	// Documentation (3fff::/20) per RFC 9637
	if b[0] == 0x3f && b[1] == 0xff && b[2]&0xf0 == 0x00 {
		return "Documentation (3fff::/20)"
	}
	// Unique Local Address (fc00::/7)
	if b[0]&0xfe == 0xfc {
		return "Unique Local Address (ULA, fc00::/7)"
	}
	// Link-Local (fe80::/10)
	if b[0] == 0xfe && b[1]&0xc0 == 0x80 {
		return "Link-Local (fe80::/10)"
	}
	// Multicast (ff00::/8) with scope
	if b[0] == 0xff {
		scope := b[1] & 0x0f
		scopeStr := ""
		switch scope {
		case 0x01:
			scopeStr = "Interface-Local"
		case 0x02:
			scopeStr = "Link-Local"
		case 0x04:
			scopeStr = "Admin-Local"
		case 0x05:
			scopeStr = "Site-Local"
		case 0x08:
			scopeStr = "Organization-Local"
		case 0x0e:
			scopeStr = "Global"
		default:
			scopeStr = fmt.Sprintf("Unknown (0x%02x)", scope)
		}
		return fmt.Sprintf("Multicast (ff00::/8), Scope: %s", scopeStr)
	}
	// Global Unicast (2000::/3)
	if b[0]&0xe0 == 0x20 {
		return "Global Unicast (2000::/3)"
	}

	return "Reserved / Unknown"
}

// compressionPermutations returns all valid :: abbreviations for consecutive zero groups of length >= 2.
func compressionPermutations(ip net.IP) []string {
	b := ip.To16()
	groups := make([]string, 8)
	for i := 0; i < 8; i++ {
		groups[i] = fmt.Sprintf("%x", uint16(b[i*2])<<8|uint16(b[i*2+1]))
	}

	var results []string
	for start := 0; start < 8; start++ {
		if groups[start] != "0" {
			continue
		}
		for end := start + 2; end <= 8; end++ {
			if groups[end-1] != "0" {
				break
			}
			left := groups[:start]
			right := groups[end:]
			var s string
			switch {
			case start == 0 && end == 8:
				s = "::"
			case start == 0:
				s = "::" + strings.Join(right, ":")
			case end == 8:
				s = strings.Join(left, ":") + "::"
			default:
				s = strings.Join(left, ":") + "::" + strings.Join(right, ":")
			}
			results = append(results, s)
		}
	}

	// Deduplicate
	seen := map[string]bool{}
	var unique []string
	for _, r := range results {
		if !seen[r] {
			seen[r] = true
			unique = append(unique, r)
		}
	}
	return unique
}

// formatIPv6 parses an IPv6 address and prints all format representations.
func formatIPv6(input string) {
	ip, prefixLen, err := parseIPv6WithOptionalPrefix(input)
	if err != nil {
		log.Fatal(err)
	}

	suffix := ""
	if prefixLen >= 0 {
		suffix = fmt.Sprintf("/%d", prefixLen)
	}

	fmt.Printf("%-16s%s%s\n", "Expanded:", expandIPv6(ip), suffix)
	fmt.Printf("%-16s%s%s\n", "Compressed:", compressIPv6(ip), suffix)
	fmt.Printf("%-16s%s\n", "Uppercase:", uppercaseIPv6(ip))
	fmt.Printf("%-16s%s\n", "URL format:", urlIPv6(ip))
	fmt.Printf("%-16s%s\n", "Dotted:", dottedIPv6(ip))
	fmt.Printf("%-16s%s\n", "Binary:", binaryIPv6(ip))

	arpa, err := ipv6ToArpa(ip.String(), 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%-16s%s\n", "Reverse DNS:", arpa)
	fmt.Printf("%-16s%s\n", "Address Type:", classifyIPv6(ip))

	perms := compressionPermutations(ip)
	if len(perms) > 0 {
		fmt.Println()
		fmt.Println("Compression permutations:")
		for _, p := range perms {
			fmt.Printf("  %s\n", p)
		}
	}
}

func main() {
	prefix := flag.String("prefix", "64:ff9b::", "IPv6 prefix for synthesis. (alias: -p)")
	newPrefixLength := flag.Int("new-prefix-length", 40, "New prefix length for subnet allocation. (alias: -n)")
	outputFile := flag.String("output", "", "File to save the output subnets. (alias: -o)")
	source := flag.String("s", "", "Source address for conversion.")
	macInput := flag.String("m", "", "SLAAC IPv6 address to decode MAC from.")
	linkLocal := flag.String("local", "", "Link-local MAC or IPv6 to convert. (alias: -a)")
	nonWellKnownPrefix := flag.String("k", "64:ff9b::", "Non-well-known prefix for RFC 6052 conversion.")
	limit := flag.Int("l", 0, "Limit the number of subnets displayed.")
	countOnly := flag.Bool("count", false, "Display only the number of generated prefixes. (alias: -c)")
	ip6arpa := flag.String("ip6.arpa", "", "Generate a reverse ip6.arpa name for an IPv6 address. Uses -new-prefix-length as zone context.")
	format := flag.String("format", "", "Display all format representations of an IPv6 address. (alias: -f)")

	flag.StringVar(prefix, "p", "64:ff9b::", "Alias for -prefix")
	flag.IntVar(newPrefixLength, "n", 40, "Alias for -new-prefix-length")
	flag.StringVar(outputFile, "o", "", "Alias for -output")
	flag.BoolVar(countOnly, "c", false, "Alias for -count")
	flag.StringVar(linkLocal, "a", "", "Alias for -local")
	flag.StringVar(format, "f", "", "Alias for -format")

	flag.Parse()

	if flag.NFlag() == 0 {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *format != "" {
		formatIPv6(*format)
		return
	}

	if *macInput != "" {
		mac, err := decodeMACFromSLAAC(*macInput)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Decoded MAC address:", mac)
		return
	}

	if *linkLocal != "" {
		if ip := net.ParseIP(*linkLocal); ip != nil && ip.To16() != nil && strings.HasPrefix(*linkLocal, "fe80") {
			mac, err := linkLocalToMAC(*linkLocal)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("MAC from link-local:", mac)
		} else {
			ll, err := macToLinkLocal(*linkLocal)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Link-local address:", ll)
		}
		return
	}

	if *source != "" {
		ip := net.ParseIP(*source)
		if ip == nil {
			log.Fatalf("Invalid IP address: %s", *source)
		}
		if ip.To4() != nil {
			synthesizedAddr, err := ipv4ToSynthesized(*source, *nonWellKnownPrefix)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Converted IPv4 to synthesized IPv6:", synthesizedAddr)
		} else {
			ipv4Addr, err := synthesizedToIPv4(*source)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Converted synthesized IPv6 to IPv4:", ipv4Addr)
		}
		return
	}

	if *ip6arpa != "" {
		arpa, err := ipv6ToArpa(*ip6arpa, *newPrefixLength)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(arpa)
		return
	}

	if *countOnly {
		count, err := countSubnets(*prefix, *newPrefixLength)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Number of prefixes: %d\n", count)
		return
	}

	subnets, err := generateSubnets(*prefix, *newPrefixLength, *limit)
	if err != nil {
		log.Fatal(err)
	}

	if *outputFile != "" {
		outputFileHandle, err := os.Create(*outputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer outputFileHandle.Close()
		for _, subnet := range subnets {
			_, err := outputFileHandle.WriteString(subnet + "\n")
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Printf("Subnets saved to %s\n", *outputFile)
	} else {
		for _, subnet := range subnets {
			fmt.Println(subnet)
		}
	}
}
