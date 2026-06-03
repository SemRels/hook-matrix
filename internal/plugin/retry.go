// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package plugin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	DefaultMaxRetries = 3
	DefaultRetryDelay = 2 * time.Second
)

var retryLogWriter io.Writer = os.Stderr

// retryDo executes fn up to maxRetries+1 times on network errors or 5xx responses.
func retryDo(ctx context.Context, maxRetries int, delay time.Duration, fn func() (*http.Response, error)) (*http.Response, error) {
	if maxRetries < 0 {
		maxRetries = 0
	}
	if delay < 0 {
		delay = 0
	}

	for attempt := 0; ; attempt++ {
		resp, err := fn()
		if !shouldRetry(ctx, resp, err) {
			return resp, err
		}
		if attempt >= maxRetries {
			return resp, err
		}
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		reason := ""
		if err != nil {
			reason = err.Error()
		} else {
			reason = fmt.Sprintf("unexpected status %d", resp.StatusCode)
		}
		_, _ = fmt.Fprintf(retryLogWriter, "retry attempt %d/%d after %s: %s\n", attempt+1, maxRetries, delay, reason)

		if delay == 0 {
			continue
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func shouldRetry(ctx context.Context, resp *http.Response, err error) bool {
	if err == nil {
		return resp != nil && resp.StatusCode >= http.StatusInternalServerError
	}
	if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}
