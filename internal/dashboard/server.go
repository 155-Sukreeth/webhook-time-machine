package dashboard

import (
	"embed"
	"net/http"
	"strings"

	"github.com/155-Sukreeth/webhook-time-machine/internal/api"
)

type Server struct {
	apiHandler *api.APIHandler
	webFS      embed.FS
}

func NewServer(apiHandler *api.APIHandler, webFS embed.FS) *Server {
	return &Server{
		apiHandler: apiHandler,
		webFS:      webFS,
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// REST API v1
	mux.HandleFunc("/api/v1/health", s.apiHandler.HealthCheck)
	mux.HandleFunc("/api/v1/webhooks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s.apiHandler.ListWebhooks(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/api/v1/webhooks/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/webhooks/")
		if r.Method == http.MethodDelete {
			s.apiHandler.DeleteWebhook(w, r, id)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	})

	mux.HandleFunc("/api/v1/replays/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/replays/")
		if r.Method == http.MethodPost {
			s.apiHandler.TriggerReplay(w, r, id)
			return
		}
		if r.Method == http.MethodGet {
			s.apiHandler.GetReplayLogs(w, r, id)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Web UI Frontend
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexData, err := s.webFS.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Failed loading UI", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexData)
	})

	return mux
}
