package client

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
)

type ProxyRotator struct {
	mu      sync.Mutex
	proxies []*url.URL
	idx     int
}

func NewProxyRotator(raw []string) (*ProxyRotator, error) {
	out := make([]*url.URL, 0, len(raw))
	for _, s := range raw {
		u, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return &ProxyRotator{proxies: out}, nil
}

func (r *ProxyRotator) Next() *url.URL {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.proxies) == 0 {
		return nil
	}
	u := r.proxies[r.idx%len(r.proxies)]
	r.idx++
	return u
}

type ProxyTransport struct {
	baseClient *http.Client

	rotator *ProxyRotator
	mu      sync.Mutex
	clients map[string]*http.Client
}

func NewProxyTransportWithList(base *http.Client, proxyList []string) (*ProxyTransport, error) {
	rot, err := NewProxyRotator(proxyList)
	if err != nil {
		return nil, err
	}
	return &ProxyTransport{
		baseClient: base,
		rotator:    rot,
		clients:    make(map[string]*http.Client),
	}, nil
}

func NewProxyTransportWithRotation(base *http.Client, rotationURL string) (*ProxyTransport, error) {
	rot, err := NewProxyRotator([]string{rotationURL})
	if err != nil {
		return nil, err
	}
	return &ProxyTransport{
		baseClient: base,
		rotator:    rot,
		clients:    make(map[string]*http.Client),
	}, nil
}

func (p *ProxyTransport) Do(req *http.Request) (*http.Response, error) {
	proxyURL := p.rotator.Next()
	if proxyURL == nil {
		return p.baseClient.Do(req)
	}

	cli, err := p.clientForProxy(proxyURL)
	if err != nil {
		return nil, err
	}

	req2 := req.Clone(req.Context())

	resp, doErr := cli.Do(req2)
	if doErr != nil {
		return nil, ProxyError{Proxy: proxyURL.String(), Err: doErr}
	}
	return resp, nil
}

func (p *ProxyTransport) clientForProxy(u *url.URL) (*http.Client, error) {
	key := u.String()

	p.mu.Lock()
	if c, ok := p.clients[key]; ok {
		p.mu.Unlock()
		return c, nil
	}
	p.mu.Unlock()

	baseTr, ok := p.baseClient.Transport.(*http.Transport)
	if !ok || baseTr == nil {
		return nil, fmt.Errorf("base http client transport must be *http.Transport")
	}

	tr := baseTr.Clone()
	tr.Proxy = http.ProxyURL(u)

	// отдельный cookie jar на каждый прокси
	jar, _ := cookiejar.New(nil)

	c := &http.Client{
		Transport:     tr,
		Timeout:       p.baseClient.Timeout,
		Jar:           jar,
		CheckRedirect: p.baseClient.CheckRedirect,
	}

	p.mu.Lock()
	p.clients[key] = c
	p.mu.Unlock()

	return c, nil
}

type ProxyError struct {
	Proxy string
	Err   error
}

func (e ProxyError) Error() string { return "proxy=" + e.Proxy + " err=" + e.Err.Error() }
func (e ProxyError) Unwrap() error { return e.Err }
