package geo

import (
	"fmt"
	"io"
	"net"

	"github.com/oschwald/geoip2-golang"
	"github.com/vladislav-koval/url-shortener/internal/platform/logger"
	"go.uber.org/zap"
)

type Client struct {
	db *geoip2.Reader
}

type Geo struct {
	Country string
	City    string
}

type Resolver interface {
	Resolve(ip string) (Geo, error)
}

type ResolveCloser interface {
	Resolver
	io.Closer
}

// NoopResolver is a Resolver that resolves nothing — used when the geo database
// isn't available, so callers can keep working without nil-checking the resolver.
type NoopResolver struct{}

func (NoopResolver) Resolve(string) (Geo, error) {
	return Geo{}, nil
}

func (NoopResolver) Close() error {
	return nil
}

// NewResolver opens the geo database. If it fails and the database isn't
// marked required (config.Required), it falls back to NoopResolver instead
// of failing — the caller doesn't need to know about that trade-off.
func NewResolver(config Config, log *logger.Logger) (ResolveCloser, error) {
	db, err := geoip2.Open(config.FilePath)
	if err != nil {
		if config.Required {
			return nil, fmt.Errorf("failed to open geo database: %w", err)
		}

		log.Warn("GEO database unavailable, click geolocation disabled", zap.Error(err))
		return NoopResolver{}, nil
	}

	return &Client{
		db: db,
	}, nil
}

func (gc *Client) Resolve(ip string) (Geo, error) {
	ipData := net.ParseIP(ip)
	if ipData == nil {
		return Geo{}, fmt.Errorf("invalid IP")
	}
	record, err := gc.db.City(ipData)
	if err != nil {
		return Geo{}, fmt.Errorf("failed to lookup IP: %w", err)
	}

	return Geo{
		Country: record.Country.IsoCode,
		City:    record.City.Names["en"],
	}, nil
}

func (gc *Client) Close() error {
	return gc.db.Close()
}
