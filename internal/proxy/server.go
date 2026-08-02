package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/155-Sukreeth/webhook-time-machine/internal/models"
	"github.com/155-Sukreeth/webhook-time-machine/internal/storage"
)

type Server struct {
	targetURL  *url.URL
	store      *storage.Storage
	httpClient *http.Client
}

func NewServer(targetURL string, store *storage.Storage) (*Server, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target forward URL: %w", err)
	}

	return &Server{
		targetURL: parsed,
		store:     store,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := uuid.New().String()
	startTime := time.Now()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request payload", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	reqHeaders := make(map[string]string)
	for k, vv := range r.Header {
		reqHeaders[k] = strings.Join(vv, ", ")
	}

	queryParams := make(map[string]string)
	for k, vv := range r.URL.Query() {
		queryParams[k] = strings.Join(vv, ", ")
	}

	webhookReq := &models.WebhookRequest{
		ID:              reqID,
		Timestamp:       startTime,
		Method:          r.Method,
		URL:             r.URL.String(),
		Path:            r.URL.Path,
		QueryParameters: queryParams,
		Headers:         reqHeaders,
		Body:            string(bodyBytes),
		ResponseHeaders: make(map[string]string),
		CreatedAt:       startTime,
	}

	destURL := *s.targetURL
	destURL.Path = r.URL.Path
	destURL.RawQuery = r.URL.RawQuery

	outReq, err := http.NewRequestWithContext(r.Context(), r.Method, destURL.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		s.logAndRespondError(w, webhookReq, "Proxy request creation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for k, vv := range r.Header {
		for _, v := range vv {
			outReq.Header.Add(k, v)
		}
	}

	fwdStart := time.Now()
	resp, err := s.httpClient.Do(outReq) // #nosec G704 -- Intentional proxy forwarding behavior
	duration := time.Since(fwdStart).Milliseconds()
	webhookReq.DurationMs = duration

	if err != nil {
		s.logAndRespondError(w, webhookReq, "Target endpoint connection error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		respBody = []byte(fmt.Sprintf("Error reading response body: %v", err))
	}
	webhookReq.ResponseStatusCode = resp.StatusCode
	webhookReq.ResponseBody = string(respBody)

	respHeaders := make(map[string]string)
	for k, vv := range resp.Header {
		respHeaders[k] = strings.Join(vv, ", ")
		w.Header()[k] = vv
	}
	webhookReq.ResponseHeaders = respHeaders

	if err := s.store.SaveRequest(r.Context(), webhookReq); err != nil {
		log.Printf("[WTM WARN] Failed to save request to storage: %v", err)
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}

func (s *Server) logAndRespondError(w http.ResponseWriter, req *models.WebhookRequest, errMsg string, code int) {
	req.ResponseStatusCode = code
	req.ResponseBody = fmt.Sprintf(`{"error": %q}`, errMsg)
	if err := s.store.SaveRequest(context.Background(), req); err != nil {
		log.Printf("[WTM WARN] Failed to save error request to storage: %v", err)
	}
	http.Error(w, errMsg, code)
}

