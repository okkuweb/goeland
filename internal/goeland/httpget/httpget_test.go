package httpget

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func resetRequestDelayForTest() {
	requestDelayMu.Lock()
	lastRequestAt = time.Time{}
	requestDelayMu.Unlock()
}

func TestRequestDelayAppliesBetweenRequests(t *testing.T) {
	resetRequestDelayForTest()
	viper.Set("request-delay", "30ms")
	t.Cleanup(func() {
		viper.Set("request-delay", time.Duration(0))
		resetRequestDelayForTest()
	})

	var mu sync.Mutex
	var hits []time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		hits = append(hits, time.Now())
		mu.Unlock()
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := *server.Client()
	if _, err := GetHTTPRessourceGeneric(server.URL, client); err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if _, err := GetHTTPRessourceGeneric(server.URL, client); err != nil {
		t.Fatalf("second request failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(hits) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(hits))
	}
	if elapsed := hits[1].Sub(hits[0]); elapsed < 25*time.Millisecond {
		t.Fatalf("expected delayed second request, elapsed %s", elapsed)
	}
}
