// Copyright 2025 - MinIO, Inc. All rights reserved.
// Use of this source code is governed by the AGPLv3
// license that can be found in the LICENSE file.

package https

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// errDown is returned by probes and round trips for a host that is down.
var errDown = errors.New("host is down")

// roundTripFunc is an http.RoundTripper that reports which hosts it
// was asked for.
type roundTripFunc func(*http.Request) (*http.Response, error)

// RoundTrip calls f.
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// Host must not return a suspended host while other hosts are available.
func TestLoadBalancerHostSkipsSuspendedHost(t *testing.T) {
	lb := &LoadBalancer{
		Hosts:         []string{"a", "b", "c"},
		Probe:         func(context.Context, string) error { return errDown },
		ProbeInterval: time.Hour,
	}
	lb.suspend("b")

	for i := 0; i < 100; i++ {
		host, err := lb.Host()
		if err != nil {
			t.Fatalf("Host: %v", err)
		}
		if host == "b" {
			t.Fatal("Host returned the suspended host 'b'")
		}
	}
}

// Host must still return a host once every host is suspended, so that a
// client keeps making progress during a full outage.
func TestLoadBalancerHostWithAllHostsSuspended(t *testing.T) {
	lb := &LoadBalancer{
		Hosts:         []string{"a", "b", "c"},
		Probe:         func(context.Context, string) error { return errDown },
		ProbeInterval: time.Hour,
	}
	for _, host := range lb.Hosts {
		lb.suspend(host)
	}

	host, err := lb.Host()
	if err != nil {
		t.Fatalf("Host: %v", err)
	}
	if host == "" {
		t.Fatal("Host returned no host while all hosts are suspended")
	}
}

// A host that does not respond must stay suspended even after Timeout
// elapsed. This is the behavior probing exists for: re-admitting an
// unreachable host on a timer makes every client pay its dial timeout
// again.
func TestLoadBalancerProbeKeepsUnresponsiveHostSuspended(t *testing.T) {
	lb := &LoadBalancer{
		Hosts:         []string{"a", "b"},
		Timeout:       time.Millisecond,
		Probe:         func(context.Context, string) error { return errDown },
		ProbeInterval: time.Millisecond,
	}
	lb.suspend("a")

	// Well past Timeout, which must not re-admit a host that a Probe
	// decides about.
	time.Sleep(20 * time.Millisecond)

	for i := 0; i < 100; i++ {
		if host, _ := lb.Host(); host != "b" {
			t.Fatalf("Host = %q, want the only host that responds", host)
		}
	}
}

// A suspended host must be re-admitted once, and only once, a probe of it
// succeeds. Probing must then stop, so that it does not outlive the
// outage.
func TestLoadBalancerProbeReadmitsHostOnlyOnceItResponds(t *testing.T) {
	live := make(chan struct{})
	probed := make(chan string, 1024)

	lb := &LoadBalancer{
		Hosts:         []string{"a", "b"},
		ProbeInterval: time.Millisecond,
	}
	lb.Probe = func(_ context.Context, host string) error {
		probed <- host
		select {
		case <-live:
			return nil
		default:
			return errDown
		}
	}
	lb.suspend("a")

	// For as long as 'a' does not respond, it must not be re-admitted,
	// however often it is probed.
	for i := 0; i < 5; i++ {
		if host := <-probed; host != "a" {
			t.Fatalf("probed host = %q, want the suspended host 'a'", host)
		}
		if host, _ := lb.Host(); host != "b" {
			t.Fatalf("Host = %q, want the only host that is not suspended", host)
		}
	}

	close(live)
	waitFor(t, "'a' to be re-admitted once it responds", func() bool {
		return !lb.suspended("a")
	})

	// Once no host is suspended anymore, probing must stop so that it
	// does not outlive the outage.
	waitFor(t, "probing to stop", func() bool {
		lb.mu.RLock()
		defer lb.mu.RUnlock()
		return !lb.probing
	})
}

// Without a Probe, Timeout keeps its previous meaning and re-admits a
// suspended host once it elapsed.
func TestLoadBalancerTimeoutReadmitsHostWithoutProbe(t *testing.T) {
	lb := &LoadBalancer{
		Hosts:   []string{"a", "b"},
		Timeout: 10 * time.Millisecond,
	}
	lb.suspend("a")

	if host, _ := lb.Host(); host != "b" {
		t.Fatalf("Host = %q, want the only host that is not suspended", host)
	}

	waitFor(t, "'a' to be re-admitted after the timeout elapsed", func() bool {
		for i := 0; i < 100; i++ {
			if host, _ := lb.Host(); host == "a" {
				return true
			}
		}
		return false
	})
}

// RoundTrip must retry a failed request with another host, suspend the
// host it failed for, and start probing that host.
func TestLoadBalancerRoundTripSuspendsFailedHost(t *testing.T) {
	probed := make(chan string, 1024)

	lb := &LoadBalancer{
		Hosts: []string{"a", "b"},
		RoundTripper: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host == "a" {
				return nil, errDown
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}),
		ProbeInterval: time.Millisecond,
	}
	lb.Probe = func(_ context.Context, host string) error {
		probed <- host
		return errDown
	}

	resp, err := lb.RoundTrip(&http.Request{URL: &url.URL{Scheme: "https", Host: "a"}})
	if err != nil {
		t.Fatalf("RoundTrip did not retry the request with a different host: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if !lb.suspended("a") {
		t.Fatal("RoundTrip did not suspend the host its request failed for")
	}
	if host := <-probed; host != "a" {
		t.Fatalf("probed host = %q, want the suspended host 'a'", host)
	}
}

// suspended reports whether host is currently suspended. It exists for
// tests only; production code consults isSuspended while holding lb.mu.
//
//nolint:unparam
func (lb *LoadBalancer) suspended(host string) bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	_, ok := lb.timeout[host]
	return ok
}

// waitFor polls cond until it holds, and fails the test if it does not
// within a fixed timeout.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()

	const Timeout = 10 * time.Second
	for deadline := time.Now().Add(Timeout); time.Now().Before(deadline); {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("timed out after %v waiting for %s", Timeout, what)
}

// A dial timeout matches context.DeadlineExceeded because net.Dialer
// implements its Timeout with a context deadline. It is the error an
// unreachable host produces, so it must suspend that host.
func TestLoadBalancerRoundTripSuspendsHostOnTimeoutWithLiveContext(t *testing.T) {
	lb := &LoadBalancer{
		Hosts: []string{"a", "b"},
		RoundTripper: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host == "a" {
				return nil, context.DeadlineExceeded
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}),
		Probe:         func(context.Context, string) error { return errDown },
		ProbeInterval: time.Hour,
	}

	resp, err := lb.RoundTrip(&http.Request{URL: &url.URL{Scheme: "https", Host: "a"}})
	if err != nil {
		t.Fatalf("RoundTrip did not retry the request with a different host: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if !lb.suspended("a") {
		t.Fatal("RoundTrip did not suspend the host that timed out")
	}
}

// An expired request context means the caller is gone, not that the host
// is unhealthy. RoundTrip must neither suspend the host nor retry.
func TestLoadBalancerRoundTripKeepsHostOnCallerTimeout(t *testing.T) {
	hosts := make(chan string, 1024)
	lb := &LoadBalancer{
		Hosts: []string{"a", "b"},
		RoundTripper: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			hosts <- req.URL.Host
			return nil, context.DeadlineExceeded
		}),
		Probe:         func(context.Context, string) error { return errDown },
		ProbeInterval: time.Hour,
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	req := (&http.Request{URL: &url.URL{Scheme: "https", Host: "a"}}).WithContext(ctx)
	if _, err := lb.RoundTrip(req); err == nil {
		t.Fatal("RoundTrip did not return the caller's error")
	}
	if lb.suspended("a") {
		t.Fatal("RoundTrip suspended a host because the caller's context expired")
	}
	if len(hosts) != 1 {
		t.Fatalf("RoundTrip sent %d requests, want 1 - it must not retry for a caller that is gone", len(hosts))
	}
}
