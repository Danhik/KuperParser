package client

import (
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"
)

func NewHTTPClient(timeout time.Duration) *http.Client {
	jar, _ := cookiejar.New(nil)

	tr := &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,

		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,

		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,

		ForceAttemptHTTP2: true,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
		Jar:       jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			prev := via[len(via)-1]
			log.Printf("[REDIRECT] %s -> %s (hops=%d)", prev.URL.String(), req.URL.String(), len(via))

			if jar != nil {
				ck := jar.Cookies(req.URL)
				names := make([]string, 0, len(ck))
				for _, c := range ck {
					names = append(names, c.Name)
				}
				log.Printf("[COOKIES] for %s: %v", req.URL.Host, names)
			}
			return nil
		},
	}

}

type HTTPTransport struct {
	Client *http.Client
}

func (h *HTTPTransport) Do(req *http.Request) (*http.Response, error) {
	return h.Client.Do(req)
}
