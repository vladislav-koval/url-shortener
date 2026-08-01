package request

import (
	"net"
	"net/http"
)

// GetClientIP TODO: доработать когда появится прокси сервер
func GetClientIP(r *http.Request) *string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil
	}

	parsedIP := net.ParseIP(host)
	if parsedIP != nil {
		if ip4 := parsedIP.To4(); ip4 != nil {
			host = ip4.String()
		}
	}

	ip := host
	return &ip
}
