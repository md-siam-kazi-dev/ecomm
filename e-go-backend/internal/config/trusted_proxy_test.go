package config

import "testing"

func TestParsePrefixes(t *testing.T) {
	t.Run("accepts CIDRs and bare addresses", func(t *testing.T) {
		got, err := parsePrefixes([]string{"10.0.0.0/8", " 192.168.1.5 ", "", "2001:db8::/32"})
		if err != nil {
			t.Fatalf("parsePrefixes: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("expected 3 prefixes (blank entry skipped), got %d: %v", len(got), got)
		}
		if got[0].String() != "10.0.0.0/8" {
			t.Fatalf("prefix[0] = %s", got[0])
		}
		// A bare address becomes a single-host prefix.
		if got[1].String() != "192.168.1.5/32" {
			t.Fatalf("prefix[1] = %s, want 192.168.1.5/32", got[1])
		}
		if got[2].String() != "2001:db8::/32" {
			t.Fatalf("prefix[2] = %s", got[2])
		}
	})

	t.Run("unmaps a bare IPv4-mapped address", func(t *testing.T) {
		got, err := parsePrefixes([]string{"::ffff:10.1.2.3"})
		if err != nil {
			t.Fatalf("parsePrefixes: %v", err)
		}
		if len(got) != 1 || got[0].String() != "10.1.2.3/32" {
			t.Fatalf("got %v, want [10.1.2.3/32]", got)
		}
	})

	t.Run("rejects malformed entries loudly", func(t *testing.T) {
		if _, err := parsePrefixes([]string{"not-a-cidr"}); err == nil {
			t.Fatal("expected an error for a malformed entry")
		}
	})
}
