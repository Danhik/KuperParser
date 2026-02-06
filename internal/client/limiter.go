package client

import "net/http"

type Limiter struct {
	sem chan struct{}
}

func NewLimiter(concurrency int) *Limiter {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &Limiter{sem: make(chan struct{}, concurrency)}
}

func (l *Limiter) Acquire() { l.sem <- struct{}{} }
func (l *Limiter) Release() { <-l.sem }

type LimitedTransport struct {
	Base    Transport
	Limiter *Limiter
}

func (t *LimitedTransport) Do(req *http.Request) (*http.Response, error) {
	t.Limiter.Acquire()
	defer t.Limiter.Release()
	return t.Base.Do(req)
}
