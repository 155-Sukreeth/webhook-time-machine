package models

import "time"

// Config represents full system configuration.
type Config struct {
	Port            int    `mapstructure:"port" yaml:"port" json:"port"`
	UIPort          int    `mapstructure:"ui_port" yaml:"ui_port" json:"ui_port"`
	ForwardURL      string `mapstructure:"forward_url" yaml:"forward_url" json:"forward_url"`
	DBPath          string `mapstructure:"db_path" yaml:"db_path" json:"db_path"`
	LogLevel        string `mapstructure:"log_level" yaml:"log_level" json:"log_level"`
	StripSignatures bool   `mapstructure:"strip_signatures" yaml:"strip_signatures" json:"strip_signatures"`
}

// WebhookRequest represents a captured incoming webhook.
type WebhookRequest struct {
	ID                 string            `json:"id"`
	Timestamp          time.Time         `json:"timestamp"`
	Method             string            `json:"method"`
	URL                string            `json:"url"`
	Path               string            `json:"path"`
	QueryParameters    map[string]string `json:"query_parameters"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body"`
	ResponseStatusCode int               `json:"response_status_code"`
	ResponseBody       string            `json:"response_body"`
	ResponseHeaders    map[string]string `json:"response_headers"`
	DurationMs         int64             `json:"duration_ms"`
	IsReplay           bool              `json:"is_replay"`
	ParentID           *string           `json:"parent_id,omitempty"`
	ReplayCount        int               `json:"replay_count"`
	Tags               []string          `json:"tags"`
	UserNotes          string            `json:"user_notes"`
	CreatedAt          time.Time         `json:"created_at"`
}

// ReplayLog records an individual replay execution.
type ReplayLog struct {
	ID                 string            `json:"id"`
	OriginalRequestID  string            `json:"original_request_id"`
	Timestamp          time.Time         `json:"timestamp"`
	TargetURL          string            `json:"target_url"`
	Method             string            `json:"method"`
	HeadersSent        map[string]string `json:"headers_sent"`
	BodySent           string            `json:"body_sent"`
	ResponseStatusCode int               `json:"response_status_code"`
	ResponseBody       string            `json:"response_body"`
	ResponseHeaders    map[string]string `json:"response_headers"`
	DurationMs         int64             `json:"duration_ms"`
	Error              string            `json:"error,omitempty"`
}

// RequestFilter specifies query criteria for listing webhooks.
type RequestFilter struct {
	Query      string     `json:"query"`
	Method     string     `json:"method"`
	StatusCode int        `json:"status_code"`
	Tag        string     `json:"tag"`
	IsReplay   *bool      `json:"is_replay"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

// ReplayRequestPayload represents customizable options when triggering a replay.
type ReplayRequestPayload struct {
	TargetURL string            `json:"target_url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
}

// APIResponse represents standard JSON API envelope response.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
