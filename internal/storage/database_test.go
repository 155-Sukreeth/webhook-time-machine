package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/155-Sukreeth/webhook-time-machine/internal/models"
	"github.com/155-Sukreeth/webhook-time-machine/internal/storage"
)

func TestStorage_SaveAndRetrieve(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wtm-storage-test-*")
	if err != nil {
		t.Fatalf("failed temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	store, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("failed init storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("failed init schema: %v", err)
	}

	req := &models.WebhookRequest{
		ID:                 "req-001",
		Method:             "POST",
		URL:                "http://localhost:8080/stripe/webhook",
		Path:               "/stripe/webhook",
		Headers:            map[string]string{"Content-Type": "application/json"},
		Body:               `{"event": "payment_intent.succeeded"}`,
		ResponseStatusCode: 200,
		ResponseBody:       `{"received": true}`,
		DurationMs:         32,
	}

	if err := store.SaveRequest(ctx, req); err != nil {
		t.Fatalf("failed saving request: %v", err)
	}

	fetched, err := store.GetRequestByID(ctx, "req-001")
	if err != nil {
		t.Fatalf("failed fetching request: %v", err)
	}
	if fetched == nil || fetched.Method != "POST" {
		t.Errorf("mismatched record fetched: %+v", fetched)
	}

	list, total, err := store.ListRequests(ctx, models.RequestFilter{Query: "payment_intent"})
	if err != nil || total != 1 || len(list) != 1 {
		t.Errorf("expected 1 result, got total %d, list len %d", total, len(list))
	}
}
