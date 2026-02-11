#!/bin/bash

echo "Testing IPv4 to IPv6 conversion..."
./ipv6utils -s 100.64.1.1

echo "Testing IPv6 to IPv4 conversion..."
./ipv6utils -s 64:ff9b::c0a8:101

#echo "Testing SLAAC MAC address decoding..."
#./ipv6utils -m 3fff:0::0200:5eff:fe00:5325

echo "Testing subnet generation..."
./ipv6utils -p 3fff:0::/32 -n 40 -l 5

echo "Testing prefix count..."
./ipv6utils -p 3fff:0::/32 -n 40 -c

echo "Testing output to file..."
./ipv6utils -p 3fff:0::/32 -n 36 -o subnets.txt
cat subnets.txt

echo "Testing link MAC to local decoder..."
./ipv6utils -local 00:11:22:33:44:55

echo "Testing link local to MAC decoder..."
./ipv6utils -local fe80::0211:22ff:fe33:4455

echo "Testing alias flags..."
./ipv6utils -p 3fff:0::/32 -n 40 -l 5
./ipv6utils -prefix 3fff:0::/32 -new-prefix-length 40 -l 5
./ipv6utils -a fe80::0211:22ff:fe33:4455

echo "Testing DNS PTR generation on /56 boundary..."
./ipv6utils -ip6.arpa 3fff:0:abcd::0211:22ff:fe33:4455 -n 56

echo "Testing DNS PTR generation..."
./ipv6utils -ip6.arpa 3fff:0:abcd::0211:22ff:fe33:4455 -n 0

echo "Testing IPv6 format display..."
./ipv6utils -f 2001:db8::1

echo "Testing IPv6 format display with prefix..."
./ipv6utils -format 2001:db8::1/48

echo "Testing IPv6 format display for loopback..."
./ipv6utils -f ::1

echo "All tests completed."
