package netutil

import (
	"testing"
)

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"::1", true},
		{"not-an-ip", false},
		{"999.999.999.999", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsValidIP(tt.input)
			if got != tt.want {
				t.Errorf("IsValidIP(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePorts(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{"single port", "80", []int{80}, false},
		{"multiple ports", "22,80,443", []int{22, 80, 443}, false},
		{"with spaces", "22, 80, 443", []int{22, 80, 443}, false},
		{"invalid port", "abc", nil, true},
		{"out of range", "99999", nil, true},
		{"zero port", "0", nil, true},
		{"negative port", "-1", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePorts(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePorts(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ParsePorts(%q) = %v, want %v", tt.input, got, tt.want)
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("ParsePorts(%q)[%d] = %d, want %d", tt.input, i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
