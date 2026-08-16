package request

import (
	"net"
	"net/http"
	"strings"
)

// GetClientIP prefers X-Forwarded-For (set by Caddy's reverse_proxy) over
// r.RemoteAddr, which behind a reverse proxy is the proxy's own address, not
// the real client's. Takes the last entry in the chain — that's the hop
// Caddy itself observed directly, so it can't be spoofed by the client;
// earlier entries in the header (if the client sent its own) could be faked.
func GetClientIP(r *http.Request) *string {
	host := lastForwardedFor(r)

	if host == "" {
		var err error
		host, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return nil
		}
	}

	parsedIP := net.ParseIP(host)
	if parsedIP == nil {
		return nil
	}

	if ip4 := parsedIP.To4(); ip4 != nil {
		host = ip4.String()
	} else {
		host = parsedIP.String()
	}

	return &host
}

func lastForwardedFor(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return ""
	}

	parts := strings.Split(xff, ",")
	return strings.TrimSpace(parts[len(parts)-1])
}
