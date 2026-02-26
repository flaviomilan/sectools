// Package netutil provides common network utility functions
// shared across sectools commands.
package netutil

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// IsValidIP checks whether the given string is a valid IP address.
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// GetLocalIPv4 returns the first non-loopback IPv4 address of the named interface.
func GetLocalIPv4(interfaceName string) (net.IP, error) {
	ifaceObj, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("interface %s: %w", interfaceName, err)
	}

	addrs, err := ifaceObj.Addrs()
	if err != nil {
		return nil, fmt.Errorf("addresses for %s: %w", interfaceName, err)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip4 := ipNet.IP.To4(); ip4 != nil {
				return ip4, nil
			}
		}
	}

	return nil, fmt.Errorf("no IPv4 address found on interface %s", interfaceName)
}

// ParsePorts splits a comma-separated port string into a slice of ints.
func ParsePorts(input string) ([]int, error) {
	parts := strings.Split(input, ",")
	ports := make([]int, 0, len(parts))
	for _, p := range parts {
		port, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid port %q: %w", p, err)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("port %d out of range (1-65535)", port)
		}
		ports = append(ports, port)
	}
	return ports, nil
}

// GrabBanner connects to host:port via TCP and reads the first response bytes.
func GrabBanner(host, port string, timeout time.Duration, payload string) (string, error) {
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return "", fmt.Errorf("connecting to %s: %w", address, err)
	}
	defer conn.Close()

	if payload != "" {
		if _, err := conn.Write([]byte(payload)); err != nil {
			return "", fmt.Errorf("sending payload to %s: %w", address, err)
		}
	}

	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return "", fmt.Errorf("setting deadline for %s: %w", address, err)
	}

	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("reading banner from %s: %w", address, err)
	}

	return string(buffer[:n]), nil
}

// MustExitf prints a formatted error message and exits with code 1.
func MustExitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
