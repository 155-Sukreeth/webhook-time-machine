package api

import (
	"encoding/json"
	"net/http"

	"github.com/155-Sukreeth/webhook-time-machine/internal/models"
	"github.com/155-Sukreeth/webhook-time-machine/internal/replay"
	"github.com/155-Sukreeth/webhook-time-machine/internal/storage"
)

type APIHandler struct {
	store    *storage.Storage
	executor *replay.Executor
	cfg      *models.Config
}

func NewAPIHandler(store *storage.Storage, executor *replay.Executor, cfg *models.Config) *APIHandler {
	return &APIHandler{
		store:    store,
		executor: executor,
		cfg:      cfg,
	}
}

func (a *APIHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	method := r.URL.Query().Get("method")

	filter := models.RequestFilter{
		Query:  query,
		Method: method,
		Limit:  100,
	}

	reqs, total, err := a.store.ListRequests(r.Context(), filter)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"requests": reqs,
			"total":    total,
		},
	})
}

func (a *APIHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request, id string) {
	if err := a.store.DeleteRequest(r.Context(), id); err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, models.APIResponse{Success: true})
}

func (a *APIHandler) TriggerReplay(w http.ResponseWriter, r *http.Request, id string) {
	var payload models.ReplayRequestPayload
	_ = json.NewDecoder(r.Body).Decode(&payload)

	logResult, err := a.executor.ExecuteReplay(r.Context(), id, payload, a.cfg.StripSignatures)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error(), Data: logResult})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: logResult})
}

func (a *APIHandler) GetReplayLogs(w http.ResponseWriter, r *http.Request, id string) {
	logs, err := a.store.GetReplayLogs(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: logs})
}

func (a *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: "healthy"})
}

func respondJSON(w http.ResponseWriter, status int, resp models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
