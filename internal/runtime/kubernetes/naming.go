package kubernetes

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

const maxDNS1123NameLength = 63

func sanitizeDNS1123(value string) string {
	var builder strings.Builder
	lastWasDash := false

	for _, char := range strings.ToLower(value) {
		valid := (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')
		if valid {
			builder.WriteRune(char)
			lastWasDash = false
			continue
		}

		if builder.Len() > 0 && !lastWasDash {
			builder.WriteByte('-')
			lastWasDash = true
		}
	}

	name := strings.Trim(builder.String(), "-")
	if name == "" {
		name = "predeploy"
	}
	if len(name) > maxDNS1123NameLength {
		name = strings.TrimRight(name[:maxDNS1123NameLength], "-")
	}
	if name == "" {
		return "predeploy"
	}

	return name
}

func namespaceName(prefix string, serviceName string, suffix string) string {
	base := sanitizeDNS1123(prefix + "-" + serviceName)
	suffix = sanitizeDNS1123(suffix)

	maxBaseLength := maxDNS1123NameLength - len(suffix) - 1
	if maxBaseLength < 1 {
		return sanitizeDNS1123(suffix)
	}
	if len(base) > maxBaseLength {
		base = strings.TrimRight(base[:maxBaseLength], "-")
	}

	return base + "-" + suffix
}

func uniqueSuffix() (string, error) {
	value := make([]byte, 4)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate runtime suffix: %w", err)
	}

	return hex.EncodeToString(value), nil
}

func findFreeLocalPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("find free local port: %w", err)
	}
	defer listener.Close()

	address, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("find free local port: unexpected listener address %T", listener.Addr())
	}

	return address.Port, nil
}
