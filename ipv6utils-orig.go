package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
)

// Function to check if the prefix length is a multiple of 4 (nibble boundary)
func isNibbleAligned(prefixLength int) bool {
	return prefixLength%4 == 0
}

// Generate subnets from a given IPv6 prefix and new prefix length
func generateSubnets(prefix string, newPrefixLength int) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(prefix)
	if err != nil {
		return nil, fmt.Errorf("invalid prefix: %v", err)
	}

	// Check if the new prefix length is a nibble boundary
	if !isNibbleAligned(newPrefixLength) {
		log.Println("Warning: new prefix length is not on a nibble boundary")
	}

	currentPrefixLength, _ := ipnet.Mask.Size()

	// If the new prefix length is smaller than the current one, that's not valid
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
	}

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

func main() {
	prefix := flag.String("prefix", "64:ff9b::/96", "IPv6 prefix to start from.")
	newPrefixLength := flag.Int("new-prefix-length", 40, "New prefix length for subnet allocation.")
	outputFile := flag.String("output", "", "File to save the output subnets (optional).")

	flag.StringVar(prefix, "p", "64:ff9b::/96", "IPv6 prefix to start from.")
	flag.IntVar(newPrefixLength, "n", 40, "New prefix length for subnet allocation.")
	flag.StringVar(outputFile, "o", "", "File to save the output subnets (optional).")

	flag.Parse()

	subnets, err := generateSubnets(*prefix, *newPrefixLength)
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