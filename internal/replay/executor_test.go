package replay_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/155-Sukreeth/webhook-time-machine/internal/models"
	"github.com/155-Sukreeth/webhook-time-machine/internal/replay"
	"github.com/155-Sukreeth/webhook-time-machine/internal/storage"
)

func TestExecutor_ReplaySignatureHeaderStripping(t *testing.T) {
	headersReceived := make(http.Header)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, vv := range r.Header {
			headersReceived[k] = vv
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer mockServer.Close()

	tempDir, _ := os.MkdirTemp("", "wtm-replay-test-*")
	defer os.RemoveAll(tempDir)

	store, err := storage.New(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed init storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	_ = store.InitSchema(ctx)

	req := &models.WebhookRequest{
		ID:     "req-sig-test",
		Method: "POST",
		URL:    mockServer.URL,
		Headers: map[string]string{
			"Content-Type":     "application/json",
			"Stripe-Signature": "t=123,v1=abc",
			"X-Hub-Signature":  "sha256=456",
			"User-Agent":       "GitHub-Hookshot",
		},
		Body: `{"event": "test"}`,
	}
	_ = store.SaveRequest(ctx, req)

	executor := replay.NewExecutor(store)
	logResult, err := executor.ExecuteReplay(ctx, "req-sig-test", models.ReplayRequestPayload{TargetURL: mockServer.URL}, true)

	if err != nil || logResult.ResponseStatusCode != 200 {
		t.Fatalf("failed replay execution: %v, status: %d", err, logResult.ResponseStatusCode)
	}

	if headersReceived.Get("Stripe-Signature") != "" {
		t.Errorf("expected Stripe-Signature header to be stripped")
	}
	if headersReceived.Get("X-Hub-Signature") != "" {
		t.Errorf("expected X-Hub-Signature header to be stripped")
	}
	if headersReceived.Get("User-Agent") != "GitHub-Hookshot" {
		t.Errorf("expected User-Agent header to be preserved")
	}
}
