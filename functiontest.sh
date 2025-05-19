#!/bin/bash

echo "Testing IPv4 to IPv6 conversion..."
go run ipv6utils.go -s 100.64.1.1

echo "Testing IPv6 to IPv4 conversion..."
go run ipv6utils.go -s 64:ff9b::c0a8:101

echo "Testing SLAAC MAC address decoding..."
go run ipv6utils.go -m 3fff:0::0200:5eff:fe00:5325

echo "Testing subnet generation..."
go run ipv6utils.go -p 3fff:0::/32 -n 40 -l 5

echo "Testing prefix count..."
go run ipv6utils.go -p 3fff:0::/32 -n 40 -c

echo "Testing output to file..."
go run ipv6utils.go -p 3fff:0::/32 -n 36 -o subnets.txt
cat subnets.txt

echo "Testing alias flags..."
go run ipv6utils.go -p 3fff:0::/32 -n 40 -l 5
#go run ipv6utils.go -prefix 3fff:0::/32 -new-prefix-length 40 -limit 5

echo "Testing link MAC to local decoder..."
go run ipv6utils.go -local 00:11:22:33:44:55

echo "Testing link local to MAC decoder..."
go run ipv6utils.go -local fe80::0211:22ff:fe33:4455 

echo "Testing DNS PTR generation on /56 boundary..."
go run ipv6utils.go -ip6.arpa 3fff:0:abcd::0211:22ff:fe33:4455 -n 56

echo "Testing DNS PTR generation..."
go run ipv6utils.go -ip6.arpa 3fff:0:abcd::0211:22ff:fe33:4455 -n 0

echo "All tests completed."
