package server

import (
	"fmt"
	"net"
	"testing"
)

func TestCIDRRange(t *testing.T) {
	cidrs := []string{"192.168.1.0/24", "10.0.0.0/32", }
	rangeChecker, err := NewCIDRRange(cidrs)
	if err != nil {
		t.Fatalf("Error creating CIDRRange: %v", err)
	}
	tests := []struct {
		ip       string
		expected bool
	}{
		// {"192.168.1.50", true},
		// {"192.168.1.0", true},
		// {"192.168.1.255", true},
		// {"192.168.1.128", true},
		// {"192.168.2.128", true},
		// {"192.168.3.128", true},
		// {"192.168.255.128", true},
		// {"192.162.1.128", true},
		// {"10.5.5.5", true},
		// {"172.16.0.1", true},
		{"10.0.0.0", true},



		// {"192.168.1.50", true},
		// {"192.168.1.0", true},
		// {"192.168.1.255", true},
		// {"192.168.1.128", true},
		// {"192.168.2.128", false},
		// {"192.168.3.128", false},
		// {"192.168.255.128", false},
		// {"192.162.1.128", false},
		// {"10.5.5.5", true},
		// {"172.16.0.1", false},
		// {"8.8.8.8", false},
	}

	for _, test := range tests {
		ip := net.ParseIP(test.ip)
		result := rangeChecker.Contains(ip)
		if result != test.expected {
			t.Errorf("IP %s in range: %v (expected: %v)", test.ip, result, test.expected)
		}
	}
}

func TestCIDRRang2e(t *testing.T) {
	cidrs := []string{  "172.16.0.0/12",}
	rangeChecker, err := NewCIDRRange(cidrs)
	if err != nil {
		t.Fatalf("Error creating CIDRRange: %v", err)
	}
	tests := []struct {
		ip       string
		expected bool
	}{
		// {"192.168.1.50", true},
		// {"10.5.5.5", true},
		// {"172.16.0.1", false},
		// {"8.8.8.8", false},
		{"172.32.0.1", false},
	}

	for _, test := range tests {
		ip := net.ParseIP(test.ip)
		result := rangeChecker.Contains(ip)
		if result != test.expected {
			t.Errorf("IP %s in range: %v (expected: %v)", test.ip, result, test.expected)
		}
	}
}




func TestCIDRRange2(t *testing.T) {
	cidrs := []string{
		"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12", "8.8.8.0/24", "1.1.1.1/32",
	}
	rangeChecker, err := NewCIDRRange(cidrs)
	if err != nil {
		t.Fatalf("Error creating CIDRRange: %v", err)
	}

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.50", true},
		{"10.5.5.5", true},
		{"172.16.5.5", true},
		{"172.31.255.255", true},
		{"172.32.0.1", false},
		{"8.8.8.8", true},
		{"8.8.4.4", false},
		{"192.168.2.1", false},
		{"0.0.0.0", false},
		{"0.1.0.1", false},
		{"1.1.1.1", true},
		{"1.1.1.2", false},
		{"1.1.2.1", false},
		{"1.2.1.2", false},
		{"1.1.2.2", false},
	}

	// Adding 400+ IP test cases
	for i := 1; i <= 250; i++ {
		tests = append(tests, struct {
			ip       string
			expected bool
		}{
			ip:       fmt.Sprintf("10.0.0.%d", i),
			expected: true,
		})
	}

	for i := 1; i <= 100; i++ {
		tests = append(tests, struct {
			ip       string
			expected bool
		}{
			ip:       fmt.Sprintf("192.168.1.%d", i),
			expected: true,
		})
	}

	for i := 1; i <= 100; i++ {
		tests = append(tests, struct {
			ip       string
			expected bool
		}{
			ip:       fmt.Sprintf("172.16.10.%d", i),
			expected: true,
		})
	}

	for _, test := range tests {
		ip := net.ParseIP(test.ip)
		if ip == nil {
			t.Errorf("Invalid IP format: %s", test.ip)
			continue
		}
		result := rangeChecker.Contains(ip)
		if result != test.expected {
			t.Errorf("IP %s in range: %v (expected: %v)", test.ip, result, test.expected)
		}
	}
}

func TestCIDRRange3(t *testing.T) {
	cidrs := []string{

		"192.168.0.0/16",
		"192.127.0.0/16",
		"192.127.0.0/32",
		"192.157.0.0/16",
		"172.16.0.0/16",
		"8.8.8.0/24",
		"1.1.1.1/32",
		"10.0.0.0/8",
		"10.0.0.0/8",
		
	}
	rangeChecker, err := NewCIDRRange(cidrs)
	if err != nil {
		t.Fatalf("Error creating CIDRRange: %v", err)
	}

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.0", true},
		{"192.168.0.0", true},
		{"192.168.5.255", true},
		{"192.157.5.255", true},
		{"192.168.0.0", true},
		{"192.168.0.0", true},
		{"192.161.1.0", false},
		{"192.161.5.0", false},
		{"192.161.5.4", false},
		{"192.161.5.4", false},
		{"192.167.5.4", false},
		{"192.167.5.4", false},
		{"192.167.5.4", false},
		{"192.166.5.4", false},
		{"192.167.5.4", false},

		{"192.127.5.4", true},
		{"192.127.5.4", true},
		{"192.127.5.4", true},
		{"192.127.5.4", true},
		{"192.127.5.4", true},


		{"0.167.5.4", false},
		{"0.166.5.4", false},
	}

	// Adding more tricky IP test cases
	

	for _, test := range tests {
		ip := net.ParseIP(test.ip)
		if ip == nil {
			t.Errorf("Invalid IP format: %s", test.ip)
			continue
		}
		result := rangeChecker.Contains(ip)
		if result != test.expected {
			t.Errorf("IP %s in range: %v (expected: %v)", test.ip, result, test.expected)
		}
	}
}

func TestCIDRRange5(t *testing.T) {
	cidrs := []string{
		"192.168.0.0/16",      // Range 1
		"192.127.0.0/16",      // Range 2
		"192.127.0.0/32",      // Range 3 (Single IP)
		"192.157.0.0/16",      // Range 4
		"172.16.0.0/16",       // Range 5
		"8.8.8.0/24",          // Range 6 (Google DNS)
		"1.1.1.1/32",          // Range 7 (Single IP)
		"10.0.0.0/8",          // Range 8
		"10.5.5.0/24",         // Range 9 (Non-contiguous block)
		"192.100.0.0/16",      // Range 10 (Overlapping Range)
		"192.168.0.0/24",      // Range 11 (Smaller subnet within Range 1)
		"0.0.0.0/0",           // Range 12 (Matches all IPs)
		"10.255.255.255/32",    // Range 13 (Single IP in the 10.0.0.0/8 block)
		"10.5.5.5/32",         // Range 14 (Single IP in the 10.5.5.0/24 block)
		"172.16.254.254/32",   // Range 15 (A single address from the 172.16.0.0/16 range)
		"8.8.8.8/32",          // Range 16 (Single IP in the 8.8.8.0/24 range)
		"255.255.255.255/32",   // Range 17 (Edge case, broadcast IP)
		"192.168.5.0/24",      // Range 18 (Another smaller subnet, outside main CIDR)
		"172.31.255.255/32",    // Range 19 (Edge case from the 172.16.0.0/16 block)
		"172.32.0.0/16",       // Range 20 (CIDR Range outside 172.16.0.0/16)
	}

	rangeChecker, err := NewCIDRRange(cidrs)
	if err != nil {
		t.Fatalf("Error creating CIDRRange: %v", err)
	}

	tests := []struct {
		ip       string
		expected bool
	}{
		// Inside CIDR ranges
		{"192.168.1.0", true},    // Inside 192.168.0.0/16
		{"192.168.0.0", true},    // Inside 192.168.0.0/16
		{"192.168.5.255", true},  // Inside 192.168.0.0/16
		{"192.157.5.255", true},  // Inside 192.157.0.0/16
		{"192.127.5.4", true},    // Inside 192.127.0.0/16
		{"8.8.8.8", true},        // Inside 8.8.8.0/24
		{"10.5.5.5", true},       // Inside 10.5.5.0/24
		{"10.5.5.255", true},     // Inside 10.5.5.0/24
		{"10.255.255.255", true}, // Inside 10.0.0.0/8 (edge case)
		{"1.1.1.1", true},       // Inside 1.1.1.1/32

		// Non-matching edge cases
		{"192.161.1.0", true},    // Outside any range
		{"192.161.5.0", true},    // Outside any range
		{"192.167.5.4", true},    // Outside any range
		{"192.127.5.4", true},    // Outside any range
		{"192.168.5.4", true},    // Outside 192.168.0.0/16
		{"255.255.255.255", true}, // Outside any range (high-end edge case)
		{"10.4.4.4", true},       // Outside 10.0.0.0/8
		{"172.32.0.0", true},     // Outside 172.16.0.0/16
		{"0.0.0.0", true},         // Inside 0.0.0.0/0

		// Testing minimal deviations
		{"192.168.1.1", true},      // Inside 192.168.0.0/16
		{"192.127.0.1", true},      // Inside 192.127.0.0/16
		{"192.127.0.2", true},      // Inside 192.127.0.0/16
		{"172.16.0.2", true},       // Inside 172.16.0.0/16
		{"10.0.0.1", true},         // Inside 10.0.0.0/8

		// Testing very small CIDRs
		{"10.255.255.255", true},   // Inside 10.0.0.0/8
		{"10.0.0.0", true},         // Inside 10.0.0.0/8 (starting IP)

		// Testing for CIDR Range with a single IP (e.g., /32)
		{"10.5.5.5", true},         // Inside 10.5.5.0/24
		{"1.1.1.1", true},          // Inside 1.1.1.1/32
		{"10.255.255.255", true},   // Inside 10.255.255.255/32
		{"172.31.255.255", true},   // Inside 172.31.255.255/32

		// Testing Edge/High Value IPs
		{"192.168.255.255", true},  // Last IP in 192.168.0.0/16
		{"192.168.0.1", true},      // First IP in 192.168.0.0/16
		{"192.168.0.255", true},    // Last IP in 192.168.0.0/16
		{"192.127.0.0", true},      // First IP in 192.127.0.0/16
		{"192.127.255.255", true},  // Last IP in 192.127.0.0/16
		{"255.255.255.255", true}, // Outside any range (high-end edge case)

		// Testing specific cases with overlapping ranges
		{"10.5.5.5", true},         // Inside 10.5.5.0/24
		{"10.255.255.255", true},   // Inside 10.0.0.0/8 (edge)
		{"10.0.0.0", true},         // Inside 10.0.0.0/8 (starting IP)

		// Testing for broadcast address
		{"255.255.255.255", true}, // Outside any range
	}

	for _, test := range tests {
		ip := net.ParseIP(test.ip)
		if ip == nil {
			t.Errorf("Invalid IP format: %s", test.ip)
			continue
		}
		result := rangeChecker.Contains(ip)
		if result != test.expected {
			t.Errorf("IP %s in range: %v (expected: %v)", test.ip, result, test.expected)
		}
	}
}
