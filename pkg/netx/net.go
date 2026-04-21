package netx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LocalIP returns the local IP address.
func LocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}
	return "", fmt.Errorf("no local IP found")
}

// LocalIPs returns all local IP addresses.
func LocalIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var ips []string
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			ips = append(ips, ipNet.IP.String())
		}
	}
	return ips, nil
}

// PublicIP returns the public IP address.
func PublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf := new(strings.Builder)
	_, err = fmt.Fscan(resp.Body, buf)
	return buf.String(), err
}

// InternalIP checks if the IP is an internal IP.
func InternalIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified()
}

// IsValidPort checks if the port is valid.
func IsValidPort(port int) bool {
	return port > 0 && port <= 65535
}

// ResolveAddr resolves the address.
func ResolveAddr(addr string) (*net.TCPAddr, error) {
	return net.ResolveTCPAddr("tcp", addr)
}

// LookupHost looks up the host.
func LookupHost(host string) ([]string, error) {
	return net.LookupHost(host)
}

// LookupCNAME looks up the CNAME.
func LookupCNAME(host string) (string, error) {
	return net.LookupCNAME(host)
}

// LookupMX looks up the MX records.
func LookupMX(domain string) ([]*net.MX, error) {
	return net.LookupMX(domain)
}

// LookupTXT looks up the TXT records.
func LookupTXT(domain string) ([]string, error) {
	return net.LookupTXT(domain)
}

// DialTimeout connects to the address with timeout.
func DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

// NewTimeoutConn creates a timeout connection.
func NewTimeoutConn(conn net.Conn, readTimeout, writeTimeout time.Duration) *TimeoutConn {
	return &TimeoutConn{Conn: conn, readTimeout: readTimeout, writeTimeout: writeTimeout}
}

// TimeoutConn is a connection with timeout.
type TimeoutConn struct {
	net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (c *TimeoutConn) Read(b []byte) (n int, err error) {
	if c.readTimeout > 0 {
		c.Conn.SetReadDeadline(time.Now().Add(c.readTimeout))
	}
	return c.Conn.Read(b)
}

func (c *TimeoutConn) Write(b []byte) (n int, err error) {
	if c.writeTimeout > 0 {
		c.Conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	return c.Conn.Write(b)
}

// HTTPGet fetches URL with timeout.
func HTTPGet(urlStr string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// HTTPPost sends POST request with JSON body.
func HTTPPost(urlStr string, body any, timeout time.Duration) ([]byte, int, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

// ParseURL parses a URL.
func ParseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

// HostPort returns host:port string.
func HostPort(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// SplitHostPort splits host:port string.
func SplitHostPort(addr string) (host string, port int, err error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}
	_, err = fmt.Sscanf(portStr, "%d", &port)
	return host, port, err
}
