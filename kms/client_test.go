// Copyright 2025 - MinIO, Inc. All rights reserved.
// Use of this source code is governed by the AGPLv3
// license that can be found in the LICENSE file.

package kms

import (
	"context"
	"net/http"
	"testing"

	"github.com/lgcorzo/kms-go/kms/internal/api"
)

type recordingRoundTripper struct {
	paths chan string
}

// RoundTrip records the path it was asked for and replies with 200 OK.
func (r *recordingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.paths <- req.URL.Path
	return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
}

// A client must probe a suspended host for liveness. Without a probe,
// the load balancer re-admits a host on a timer, even one that is
// still unreachable.
func TestNewClientProbesForLiveness(t *testing.T) {
	transport := &recordingRoundTripper{paths: make(chan string, 1)}

	client, err := NewClient(&Config{
		Endpoints: []string{"kms-0.local:7373"},
		Transport: transport,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if client.lb.Probe == nil {
		t.Fatal("NewClient did not configure a probe")
	}

	if err := client.lb.Probe(context.Background(), "kms-1.local:7373"); err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if path := <-transport.paths; path != api.PathHealthLive {
		t.Fatalf("Probe requested %q, want %q", path, api.PathHealthLive)
	}
}
