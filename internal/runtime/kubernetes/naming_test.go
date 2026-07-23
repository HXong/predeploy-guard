package kubernetes

import (
	"regexp"
	"strings"
	"testing"
)

func TestSanitizeDNS1123(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "underscores and capitals", input: "Booking_Service", want: "booking-service"},
		{name: "spaces and punctuation", input: " API / Worker! ", want: "api-worker"},
		{name: "repeated separators", input: "api___worker", want: "api-worker"},
		{name: "empty result", input: "___", want: "predeploy"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := sanitizeDNS1123(test.input); got != test.want {
				t.Fatalf("sanitizeDNS1123(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestNamespaceNameIsDNS1123CompatibleAndKeepsSuffix(t *testing.T) {
	got := namespaceName(
		strings.Repeat("Long_Prefix_", 8),
		strings.Repeat("Service_Name_", 8),
		"a1b2c3d4",
	)

	if len(got) > maxDNS1123NameLength {
		t.Fatalf("namespace length = %d, want at most %d: %q", len(got), maxDNS1123NameLength, got)
	}
	if !strings.HasSuffix(got, "-a1b2c3d4") {
		t.Fatalf("namespace = %q, want unique suffix to be preserved", got)
	}
	if !regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`).MatchString(got) {
		t.Fatalf("namespace = %q, want DNS-1123 compatible name", got)
	}
}
