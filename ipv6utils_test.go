package main

import (
	"net"
	"sort"
	"testing"
)

func TestNetworkAddress(t *testing.T) {
	cases := []struct {
		name      string
		ip        string
		prefixLen int
		expect    string
	}{
		{name: "documentation /48", ip: "2001:db8::1", prefixLen: 48, expect: "2001:0db8:0000:0000:0000:0000:0000:0000"},
		{name: "link-local /64", ip: "fe80::aabb:ccff:fedd:eeff", prefixLen: 64, expect: "fe80:0000:0000:0000:0000:0000:0000:0000"},
		{name: "host route /128", ip: "::1", prefixLen: 128, expect: "0000:0000:0000:0000:0000:0000:0000:0001"},
		{name: "default route /0", ip: "2001:db8::1", prefixLen: 0, expect: "0000:0000:0000:0000:0000:0000:0000:0000"},
		{name: "IPv4-mapped /96", ip: "::ffff:192.0.2.1", prefixLen: 96, expect: "0000:0000:0000:0000:0000:ffff:0000:0000"},
		{name: "ULA /48", ip: "fd12:3456:789a:1::1", prefixLen: 48, expect: "fd12:3456:789a:0000:0000:0000:0000:0000"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tc.ip)
			}
			got := expandIPv6(networkAddress(ip.To16(), tc.prefixLen))
			if got != tc.expect {
				t.Errorf("expected %s, got %s", tc.expect, got)
			}
		})
	}
}

func TestLastAddress(t *testing.T) {
	cases := []struct {
		name      string
		ip        string
		prefixLen int
		expect    string
	}{
		{name: "documentation /48", ip: "2001:db8::1", prefixLen: 48, expect: "2001:0db8:0000:ffff:ffff:ffff:ffff:ffff"},
		{name: "link-local /64", ip: "fe80::1", prefixLen: 64, expect: "fe80:0000:0000:0000:ffff:ffff:ffff:ffff"},
		{name: "host route /128", ip: "::1", prefixLen: 128, expect: "0000:0000:0000:0000:0000:0000:0000:0001"},
		{name: "default route /0", ip: "2001:db8::1", prefixLen: 0, expect: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"},
		{name: "ULA /48", ip: "fd12:3456:789a:1::1", prefixLen: 48, expect: "fd12:3456:789a:ffff:ffff:ffff:ffff:ffff"},
		{name: "global /32", ip: "2001:db8:dead:beef::1", prefixLen: 32, expect: "2001:0db8:ffff:ffff:ffff:ffff:ffff:ffff"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tc.ip)
			}
			got := expandIPv6(lastAddress(ip.To16(), tc.prefixLen))
			if got != tc.expect {
				t.Errorf("expected %s, got %s", tc.expect, got)
			}
		})
	}
}

func TestHostSuffix(t *testing.T) {
	cases := []struct {
		name      string
		ip        string
		prefixLen int
		expect    string
	}{
		{name: "documentation /48 host ::1", ip: "2001:db8::1", prefixLen: 48, expect: "::1"},
		{name: "link-local /64 EUI-64", ip: "fe80::aabb:ccff:fedd:eeff", prefixLen: 64, expect: "::aabb:ccff:fedd:eeff"},
		{name: "host route /128 has no host bits", ip: "::1", prefixLen: 128, expect: "::"},
		{name: "network address has zero suffix", ip: "2001:db8::", prefixLen: 32, expect: "::"},
		{name: "default route /0 suffix is full address", ip: "2001:db8::1", prefixLen: 0, expect: "2001:db8::1"},
		{name: "mid-prefix host bits", ip: "fd00::1:2:3:4", prefixLen: 48, expect: "::1:2:3:4"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tc.ip)
			}
			got := hostSuffix(ip.To16(), tc.prefixLen).String()
			if got != tc.expect {
				t.Errorf("expected %s, got %s", tc.expect, got)
			}
		})
	}
}

func TestMixedNotation(t *testing.T) {
	cases := []struct {
		name   string
		ip     string
		expect string
	}{
		{name: "IPv4-mapped documentation", ip: "::ffff:192.0.2.1", expect: "::ffff:192.0.2.1"},
		{name: "IPv4-mapped RFC 1918", ip: "::ffff:10.0.0.1", expect: "::ffff:10.0.0.1"},
		{name: "IPv4-mapped all zeros", ip: "::ffff:0.0.0.0", expect: "::ffff:0.0.0.0"},
		{name: "loopback is not IPv4-mapped", ip: "::1", expect: ""},
		{name: "unspecified is not IPv4-in-IPv6", ip: "::", expect: ""},
		{name: "global unicast is not IPv4-in-IPv6", ip: "2001:db8::1", expect: ""},
		{name: "NAT64 is not IPv4-in-IPv6 mixed", ip: "64:ff9b::192.0.2.1", expect: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tc.ip)
			}
			got := mixedNotation(ip.To16())
			if got != tc.expect {
				t.Errorf("expected %q, got %q", tc.expect, got)
			}
		})
	}
}

func TestIp6Arpa(t *testing.T) {
	cases := []struct {
		name         string
		v6addr       string
		prefixLength int
		expectError  bool
		expect       string
	}{
		{
			name:         "valid address with valid prefix",
			v6addr:       "2001:db8:abcd::0211:22ff:fe33:4455",
			prefixLength: 56,
			expect:       "5.5.4.4.3.3.e.f.f.f.2.2.1.1.2.0.0.0",
		},
		{
			name:         "valid address with 0 prefix",
			v6addr:       "2001:db8:abcd::0211:22ff:fe33:4455",
			prefixLength: 0,
			expect:       "5.5.4.4.3.3.e.f.f.f.2.2.1.1.2.0.0.0.0.0.d.c.b.a.8.b.d.0.1.0.0.2.ip6.arpa.",
		},
		{
			name:         "valid address with 128 prefix",
			v6addr:       "2001:db8:abcd::0211:22ff:fe33:4455",
			prefixLength: 128,
			expect:       "",
		},
		{
			name:         "valid address with non-nibble-boundary prefix",
			v6addr:       "2001:db8:abcd::0211:22ff:fe33:4455",
			prefixLength: 122,
			expect:       "5.5",
		},
		{
			name:         "Invalid address",
			v6addr:       "xyzzy",
			prefixLength: 0,
			expectError:  true,
		},
		{
			name:         "Prefix too small",
			v6addr:       "2001:db8:abcd::0211:22ff:fe33:4455",
			prefixLength: -32,
			expectError:  true,
		},
		{
			name:         "Prefix too big",
			v6addr:       "2001:db8:abcd::0211:22ff:fe33:4455",
			prefixLength: 129,
			expectError:  true,
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.name, func(t *testing.T) {
			got, err := ipv6ToArpa(testcase.v6addr, testcase.prefixLength)
			if (err == nil) == testcase.expectError {
				t.Errorf("expected error %v, got %v", testcase.expectError, err)
			}
			if err == nil {
				if got != testcase.expect {
					t.Errorf("expected result \"%s\", got \"%s\"", testcase.expect, got)
				}
			}
		})
	}
}

func TestParseIPv6WithOptionalPrefix(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		expectIP    string
		expectPfx   int
		expectError bool
	}{
		{name: "bare address", input: "2001:db8::1", expectIP: "2001:db8::1", expectPfx: -1},
		{name: "address with prefix", input: "2001:db8::1/48", expectIP: "2001:db8::1", expectPfx: 48},
		{name: "loopback", input: "::1", expectIP: "::1", expectPfx: -1},
		{name: "loopback with prefix", input: "::1/128", expectIP: "::1", expectPfx: 128},
		{name: "full expanded address", input: "2001:0db8:0000:0000:0000:0000:0000:0001", expectIP: "2001:db8::1", expectPfx: -1},
		{name: "prefix zero", input: "2001:db8::1/0", expectIP: "2001:db8::1", expectPfx: 0},
		{name: "invalid address", input: "not-an-address", expectError: true},
		{name: "IPv4 rejected", input: "192.168.1.1", expectError: true},
		{name: "prefix too large", input: "2001:db8::1/129", expectError: true},
		{name: "prefix negative", input: "2001:db8::1/-1", expectError: true},
		{name: "empty input", input: "", expectError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip, pfx, err := parseIPv6WithOptionalPrefix(tc.input)
			if (err == nil) == tc.expectError {
				t.Errorf("expected error %v, got %v", tc.expectError, err)
			}
			if err == nil {
				if ip.String() != tc.expectIP {
					t.Errorf("expected IP %s, got %s", tc.expectIP, ip.String())
				}
				if pfx != tc.expectPfx {
					t.Errorf("expected prefix %d, got %d", tc.expectPfx, pfx)
				}
			}
		})
	}
}

func TestExpandIPv6(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "loopback", input: "::1", expect: "0000:0000:0000:0000:0000:0000:0000:0001"},
		{name: "all zeros", input: "::", expect: "0000:0000:0000:0000:0000:0000:0000:0000"},
		{name: "documentation", input: "2001:db8::1", expect: "2001:0db8:0000:0000:0000:0000:0000:0001"},
		{name: "all ones", input: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", expect: "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"},
		{name: "mixed zeros", input: "2001:db8:0:1::2", expect: "2001:0db8:0000:0001:0000:0000:0000:0002"},
		{name: "link-local", input: "fe80::1", expect: "fe80:0000:0000:0000:0000:0000:0000:0001"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.input)
			got := expandIPv6(ip)
			if got != tc.expect {
				t.Errorf("expected %s, got %s", tc.expect, got)
			}
		})
	}
}

func TestBinaryIPv6(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "loopback",
			input:  "::1",
			expect: "0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000001",
		},
		{
			name:   "all ones",
			input:  "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			expect: "1111111111111111:1111111111111111:1111111111111111:1111111111111111:1111111111111111:1111111111111111:1111111111111111:1111111111111111",
		},
		{
			name:   "documentation",
			input:  "2001:db8::1",
			expect: "0010000000000001:0000110110111000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000000:0000000000000001",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.input)
			got := binaryIPv6(ip)
			if got != tc.expect {
				t.Errorf("expected %s, got %s", tc.expect, got)
			}
		})
	}
}

func TestDottedIPv6(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "loopback",
			input:  "::1",
			expect: "0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1",
		},
		{
			name:   "documentation",
			input:  "2001:db8::1",
			expect: "2.0.0.1.0.d.b.8.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.input)
			got := dottedIPv6(ip)
			if got != tc.expect {
				t.Errorf("expected %s, got %s", tc.expect, got)
			}
		})
	}
}

func TestClassifyIPv6(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "loopback", input: "::1", expect: "Loopback (::1)"},
		{name: "unspecified", input: "::", expect: "Unspecified (::)"},
		{name: "link-local", input: "fe80::1", expect: "Link-Local (fe80::/10)"},
		{name: "ULA fd", input: "fd00::1", expect: "Unique Local Address (ULA, fc00::/7)"},
		{name: "ULA fc", input: "fc00::1", expect: "Unique Local Address (ULA, fc00::/7)"},
		{name: "documentation 2001:db8", input: "2001:db8::1", expect: "Documentation (2001:db8::/32)"},
		{name: "documentation 3fff", input: "3fff::1", expect: "Documentation (3fff::/20)"},
		{name: "multicast global", input: "ff0e::1", expect: "Multicast (ff00::/8), Scope: Global"},
		{name: "multicast link-local", input: "ff02::1", expect: "Multicast (ff00::/8), Scope: Link-Local"},
		{name: "multicast site-local", input: "ff05::1", expect: "Multicast (ff00::/8), Scope: Site-Local"},
		{name: "multicast interface-local", input: "ff01::1", expect: "Multicast (ff00::/8), Scope: Interface-Local"},
		{name: "NAT64 well-known", input: "64:ff9b::c0a8:101", expect: "NAT64 Well-Known Prefix (64:ff9b::/96)"},
		{name: "NAT64 network-specific", input: "64:ff9b:1::1", expect: "NAT64 Network-Specific (64:ff9b:1::/48)"},
		{name: "Teredo", input: "2001::1", expect: "Teredo (2001:0000::/32)"},
		{name: "6to4", input: "2002:c0a8:101::1", expect: "6to4 (2002::/16)"},
		{name: "discard", input: "100::1", expect: "Discard-Only (100::/64)"},
		{name: "IPv4-mapped", input: "::ffff:192.168.1.1", expect: "IPv4-Mapped (::ffff:0:0/96)"},
		{name: "global unicast", input: "2607:f8b0:4004:800::200e", expect: "Global Unicast (2000::/3)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.input)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tc.input)
			}
			got := classifyIPv6(ip)
			if got != tc.expect {
				t.Errorf("expected %q, got %q", tc.expect, got)
			}
		})
	}
}

func TestCompressionPermutations(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		expectCount int
		expectItems []string
	}{
		{
			name:        "no consecutive zeros",
			input:       "2001:db8:1:2:3:4:5:6",
			expectCount: 0,
			expectItems: nil,
		},
		{
			name:        "single zero group - no permutations",
			input:       "2001:db8:0:1:2:3:4:5",
			expectCount: 0,
			expectItems: nil,
		},
		{
			name:        "two consecutive zeros",
			input:       "2001:db8:0:0:1:2:3:4",
			expectCount: 1,
			expectItems: []string{"2001:db8::1:2:3:4"},
		},
		{
			name:        "three consecutive zeros gives multiple permutations",
			input:       "2001:db8:0:0:0:1:2:3",
			expectCount: 3,
			expectItems: []string{"2001:db8::0:1:2:3", "2001:db8:0::1:2:3", "2001:db8::1:2:3"},
		},
		{
			name:  "two separate zero runs",
			input: "2001:0:0:1:0:0:0:1",
			expectItems: []string{
				"2001::1:0:0:0:1",
				"2001:0:0:1::1",
				"2001:0:0:1::0:1",
				"2001:0:0:1:0::1",
			},
		},
		{
			name:        "all zeros",
			input:       "::",
			expectCount: 28,
			expectItems: []string{"::"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.input)
			if ip == nil {
				t.Fatalf("failed to parse IP: %s", tc.input)
			}
			got := compressionPermutations(ip)

			if tc.expectCount > 0 && len(got) != tc.expectCount {
				t.Errorf("expected %d permutations, got %d: %v", tc.expectCount, len(got), got)
			}

			for _, want := range tc.expectItems {
				found := false
				for _, g := range got {
					if g == want {
						found = true
						break
					}
				}
				if !found {
					sort.Strings(got)
					t.Errorf("expected permutation %q not found in %v", want, got)
				}
			}
		})
	}
}
