package request

import (
	"net"
	"net/http"
)

// GetClientIP TODO: доработать когда появится прокси сервер
func GetClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
