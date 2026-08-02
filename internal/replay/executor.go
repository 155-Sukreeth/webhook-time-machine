package replay

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/155-Sukreeth/webhook-time-machine/internal/models"
	"github.com/155-Sukreeth/webhook-time-machine/internal/storage"
)

type Executor struct {
	store      *storage.Storage
	httpClient *http.Client
}

func NewExecutor(store *storage.Storage) *Executor {
	return &Executor{
		store: store,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Known Signature Headers to strip during replay if configured.
var DefaultSignatureHeaders = []string{
	"x-hub-signature",
	"x-hub-signature-256",
	"stripe-signature",
	"x-shopify-hmac-sha256",
	"x-twilio-signature",
	"x-svix-signature",
	"x-slack-signature",
	"x-github-event-signature",
}

func (e *Executor) ExecuteReplay(ctx context.Context, origID string, payload models.ReplayRequestPayload, stripSigs bool) (*models.ReplayLog, error) {
	origReq, err := e.store.GetRequestByID(ctx, origID)
	if err != nil {
		return nil, fmt.Errorf("error reading request: %w", err)
	}
	if origReq == nil {
		return nil, fmt.Errorf("request ID %s not found", origID)
	}

	targetURL := payload.TargetURL
	if targetURL == "" {
		targetURL = origReq.URL
	}
	method := payload.Method
	if method == "" {
		method = origReq.Method
	}
	headers := payload.Headers
	if headers == nil {
		headers = origReq.Headers
	}
	body := payload.Body
	if body == "" {
		body = origReq.Body
	}

	headersToSend := make(map[string]string)
	for k, v := range headers {
		lower := strings.ToLower(k)
		if stripSigs {
			isSig := false
			for _, sig := range DefaultSignatureHeaders {
				if lower == sig {
					isSig = true
					break
				}
			}
			if isSig {
				continue
			}
		}
		headersToSend[k] = v
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("failed preparing request: %w", err)
	}

	for k, v := range headersToSend {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := e.httpClient.Do(req)
	duration := time.Since(start).Milliseconds()

	log := &models.ReplayLog{
		ID:                uuid.New().String(),
		OriginalRequestID: origID,
		Timestamp:         start,
		TargetURL:         targetURL,
		Method:            method,
		HeadersSent:       headersToSend,
		BodySent:          body,
		DurationMs:        duration,
	}

	if err != nil {
		log.Error = err.Error()
		_ = e.store.SaveReplayLog(ctx, log)
		return log, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.ResponseStatusCode = resp.StatusCode
	log.ResponseBody = string(respBody)

	rHeaders := make(map[string]string)
	for k, vv := range resp.Header {
		rHeaders[k] = strings.Join(vv, ", ")
	}
	log.ResponseHeaders = rHeaders

	if err := e.store.SaveReplayLog(ctx, log); err != nil {
		return log, fmt.Errorf("failed storing replay log: %w", err)
	}

	return log, nil
}
