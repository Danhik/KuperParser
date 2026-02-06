package client

import (
	"fmt"
	"net/http"
	"time"
)

type Transport interface {
	Do(req *http.Request) (*http.Response, error)
}

type ProxyMode string

const (
	ProxyDisabled ProxyMode = "disabled"
	ProxyList     ProxyMode = "list"
	ProxyRotation ProxyMode = "rotation"
)

type TransportConfig struct {
	Timeout     time.Duration
	Retries     int
	Workers     int
	ProxyMode   ProxyMode
	ProxyList   []string
	RotationURL string
}

func Build(baseHTTP *http.Client, cfg TransportConfig) (Transport, error) {
	var t Transport = &HTTPTransport{Client: baseHTTP}

	switch cfg.ProxyMode {
	case ProxyDisabled, "":
	case ProxyRotation:
		pt, err := NewProxyTransportWithRotation(baseHTTP, cfg.RotationURL)
		if err != nil {
			return nil, err
		}
		t = pt
	case ProxyList:
		pt, err := NewProxyTransportWithList(baseHTTP, cfg.ProxyList)
		if err != nil {
			return nil, err
		}
		t = pt
	default:
		return nil, fmt.Errorf("Не удалось получить режим прокси: %s", string(cfg.ProxyMode))
	}

	// retry layer (around proxy)
	if cfg.Retries > 0 {
		t = &RetryTransport{
			Base:       t,
			MaxRetries: cfg.Retries,
			BaseDelay:  300 * time.Millisecond,
			MaxDelay:   8 * time.Second,
		}
	}

	// limiter layer (outermost)
	if cfg.Workers > 0 {
		t = &LimitedTransport{
			Base:    t,
			Limiter: NewLimiter(cfg.Workers),
		}
	}

	return t, nil
}
