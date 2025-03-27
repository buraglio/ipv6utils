#!/bin/bash

echo "Testing IPv4 to IPv6 conversion..."
./ipv6utils -s 100.64.1.1

echo "Testing IPv6 to IPv4 conversion..."
./ipv6utils -s 64:ff9b::c0a8:101

#echo "Testing SLAAC MAC address decoding..."
#./ipv6utils -m 3ffe:0::0200:5eff:fe00:5325

echo "Testing subnet generation..."
./ipv6utils -p 3ffe:0::/32 -n 40 -l 5

echo "Testing prefix count..."
./ipv6utils -p 3ffe:0::/32 -n 40 -c

echo "Testing output to file..."
./ipv6utils -p 3ffe:0::/32 -n 36 -o subnets.txt
cat subnets.txt

echo "Testing alias flags..."
./ipv6utils -p 3ffe:0::/32 -n 40 -l 5
./ipv6utils -prefix 3ffe:0::/32 -new-prefix-length 40 -l 5

echo "All tests completed."
