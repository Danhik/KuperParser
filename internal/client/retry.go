package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

type RetryTransport struct {
	Base       Transport
	MaxRetries int

	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func (r *RetryTransport) Do(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		if err := req.Context().Err(); err != nil {
			return nil, err
		}

		curReq := req.Clone(req.Context())

		resp, err := r.Base.Do(curReq)
		if err == nil && resp != nil {
			if !shouldRetryStatus(resp.StatusCode) {
				return resp, nil
			}

			io.Copy(io.Discard, io.LimitReader(resp.Body, 32*1024))
			resp.Body.Close()

			lastErr = fmt.Errorf("retryable status=%d", resp.StatusCode)

			log.Printf("[RETRY] attempt=%d/%d status=%d url=%s",
				attempt+1, r.MaxRetries+1, resp.StatusCode, req.URL.String(),
			)

			// 429: можем подождать Retry-After
			if resp.StatusCode == http.StatusTooManyRequests {
				if d := retryAfterDelay(resp); d > 0 && attempt < r.MaxRetries {
					log.Printf("[RETRY] 429 retry-after=%s", d)
					time.Sleep(d)
					continue
				}
			}
		} else {
			// err != nil
			if err != nil && !shouldRetryError(err) {
				return nil, err
			}
			lastErr = err

			if pe := (ProxyError{}); errors.As(err, &pe) {
				log.Printf("[RETRY] attempt=%d/%d proxy=%s err=%v url=%s",
					attempt+1, r.MaxRetries+1, pe.Proxy, pe.Err, req.URL.String(),
				)
			} else {
				log.Printf("[RETRY] attempt=%d/%d err=%v url=%s",
					attempt+1, r.MaxRetries+1, err, req.URL.String(),
				)
			}
		}

		if attempt == r.MaxRetries {
			break
		}
		d := r.backoff(attempt)
		log.Printf("[RETRY] sleeping=%s before next attempt", d)
		time.Sleep(d)
	}

	return nil, lastErr
}

func shouldRetryStatus(code int) bool {
	if code == http.StatusTooManyRequests {
		return true
	}
	return code >= 500 && code <= 599
}

func shouldRetryError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var pe ProxyError
	if errors.As(err, &pe) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return false
}

func (r *RetryTransport) backoff(attempt int) time.Duration {
	d := r.BaseDelay << attempt
	if d > r.MaxDelay {
		d = r.MaxDelay
	}
	j := 0.5 + rand.Float64()
	return time.Duration(float64(d) * j)
}

func retryAfterDelay(resp *http.Response) time.Duration {
	ra := resp.Header.Get("Retry-After")
	if ra == "" {
		return 0
	}
	sec, err := strconv.Atoi(ra)
	if err != nil || sec <= 0 {
		return 0
	}
	if sec > 60 {
		sec = 60
	}
	return time.Duration(sec) * time.Second
}
