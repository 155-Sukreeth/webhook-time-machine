package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	resp, err := s.httpClient.Do(outReq)
	duration := time.Since(fwdStart).Milliseconds()
	webhookReq.DurationMs = duration

	if err != nil {
		s.logAndRespondError(w, webhookReq, "Target endpoint connection error: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	webhookReq.ResponseStatusCode = resp.StatusCode
	webhookReq.ResponseBody = string(respBody)

	respHeaders := make(map[string]string)
	for k, vv := range resp.Header {
		respHeaders[k] = strings.Join(vv, ", ")
		w.Header()[k] = vv
	}
	webhookReq.ResponseHeaders = respHeaders

	_ = s.store.SaveRequest(r.Context(), webhookReq)

	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}

func (s *Server) logAndRespondError(w http.ResponseWriter, req *models.WebhookRequest, errMsg string, code int) {
	req.ResponseStatusCode = code
	req.ResponseBody = fmt.Sprintf(`{"error": %q}`, errMsg)
	_ = s.store.SaveRequest(context.Background(), req)
	http.Error(w, errMsg, code)
}
