package main

import "testing"

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
