package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"sort"
	"strings"
	"bytes"
)

// Function to check if the prefix length is a multiple of 4 (nibble boundary)
func isNibbleAligned(prefixLength int) bool {
	return prefixLength%4 == 0
}

// Generate subnets from a given IPv6 prefix and new prefix length with optional limit
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

	subnets := []string{}
	prefixIP := ipnet.IP.Mask(ipnet.Mask)
	increment := big.NewInt(1)
	increment.Lsh(increment, uint(128-newPrefixLength))

	for i := 0; i < (1 << (newPrefixLength - currentPrefixLength)); i++ {
		subnets = append(subnets, fmt.Sprintf("%s/%d", prefixIP, newPrefixLength))
		prefixIP = addBigIntToIP(prefixIP, increment)

		if limit > 0 && len(subnets) >= limit {
			break
		}
	}

	// Sort subnets numerically by IP value
	sort.Slice(subnets, func(i, j int) bool {
		ip1 := net.ParseIP(strings.Split(subnets[i], "/")[0])
		ip2 := net.ParseIP(strings.Split(subnets[j], "/")[0])
		return bytes.Compare(ip1, ip2) < 0
	})

	return subnets, nil
}

// Adds a big integer value to an IP address
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

// Convert an RFC 6052 synthesized IPv6 address to an IPv4 address
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

// Convert an IPv4 address to an RFC 6052 synthesized IPv6 address
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

func main() {
	// Define flags with aliases
	prefix := flag.String("prefix", "64:ff9b::", "IPv6 prefix for synthesis. (alias: -p)")
	newPrefixLength := flag.Int("new-prefix-length", 40, "New prefix length for subnet allocation. (alias: -n)")
	outputFile := flag.String("output", "", "File to save the output subnets. (alias: -o)")
	source := flag.String("s", "", "Source address for conversion.")
	nonWellKnownPrefix := flag.String("k", "64:ff9b::", "Non-well-known prefix for RFC 6052 conversion.")
	limit := flag.Int("l", 0, "Limit the number of subnets displayed.")

	// Define aliases
	flag.StringVar(prefix, "p", "64:ff9b::", "Alias for -prefix")
	flag.IntVar(newPrefixLength, "n", 40, "Alias for -new-prefix-length")
	flag.StringVar(outputFile, "o", "", "Alias for -output")

	// Parse flags
	flag.Parse()

	// If no flags are provided, display help
	if flag.NFlag() == 0 {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Handle address conversion
	if *source != "" {
		ip := net.ParseIP(*source)
		if ip == nil {
			log.Fatalf("Invalid IP address: %s", *source)
		}

		if ip.To4() != nil {
			// Convert IPv4 to synthesized IPv6
			synthesizedAddr, err := ipv4ToSynthesized(*source, *nonWellKnownPrefix)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Converted IPv4 to synthesized IPv6:", synthesizedAddr)
		} else {
			// Convert synthesized IPv6 to IPv4
			ipv4Addr, err := synthesizedToIPv4(*source)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Converted synthesized IPv6 to IPv4:", ipv4Addr)
		}
		return
	}

	// Generate subnets if no conversion is requested
	subnets, err := generateSubnets(*prefix, *newPrefixLength, *limit)
	if err != nil {
		log.Fatal(err)
	}

	// Output the subnets
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
