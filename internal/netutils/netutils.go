package netutils

import (
	"net"
)

// GetIPv4Addr returns a IPv4 net.Addr for the local machine based on the network ("tcp", "udp") and port.
func GetIPv4Addr(network string, port int) net.Addr {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return fallbackAddr(network, port)
	}

	for _, addr := range addrs {
		// Check that the address is not a loopback
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			// Only use IPv4 addresses
			if ip := ipNet.IP.To4(); ip != nil {
				return makeAddr(network, ip, port)
			}
		}
	}

	return fallbackAddr(network, port)
}

func fallbackAddr(network string, port int) net.Addr {
	ip := net.ParseIP("127.0.0.1")
	return makeAddr(network, ip, port)
}

func makeAddr(network string, ip net.IP, port int) net.Addr {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return &net.TCPAddr{IP: ip, Port: port}
	case "udp", "udp4", "udp6":
		return &net.UDPAddr{IP: ip, Port: port}
	default:
		// unknown network, fallback to TCPAddr
		return &net.TCPAddr{IP: ip, Port: port}
	}
}

// IsValidPort returns true if passed in port is with the valid range and false otherwise.
func IsValidPort(port int) bool {
	return port >= 1 && port <= 65535
}

// IsValidTCPAddress checks if the given address is a valid IP:port format
func IsValidTCPAddress(addr string) bool {
	// try to resolve the TCP address to validate it
	_, err := net.ResolveTCPAddr("tcp", addr)
	return err == nil
}
