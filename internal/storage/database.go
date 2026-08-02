package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
	"github.com/155-Sukreeth/webhook-time-machine/internal/models"
	"github.com/155-Sukreeth/webhook-time-machine/internal/utils"
)

//go:embed schema.sql
var schemaFS embed.FS

type Storage struct {
	db *sql.DB
}

func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		return nil, fmt.Errorf("failed opening database at %s: %w", dbPath, err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed pinging database: %w", err)
	}
	return &Storage{db: db}, nil
}

func (s *Storage) InitSchema(ctx context.Context) error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed reading schema SQL: %w", err)
	}
	_, err = s.db.ExecContext(ctx, string(schema))
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

// Webhook Repository Implementations
func (s *Storage) SaveRequest(ctx context.Context, req *models.WebhookRequest) error {
	query := `INSERT INTO requests (id, timestamp, method, url, path, query_parameters, headers, body, response_status_code, response_body, response_headers, duration_ms, is_replay, parent_id, replay_count, tags, user_notes, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		req.ID, req.Timestamp, req.Method, req.URL, req.Path,
		utils.MapToJSON(req.QueryParameters), utils.MapToJSON(req.Headers), req.Body,
		req.ResponseStatusCode, req.ResponseBody, utils.MapToJSON(req.ResponseHeaders), req.DurationMs,
		req.IsReplay, req.ParentID, req.ReplayCount, utils.SliceToJSON(req.Tags), req.UserNotes, req.CreatedAt,
	)
	return err
}

func (s *Storage) GetRequestByID(ctx context.Context, id string) (*models.WebhookRequest, error) {
	query := `SELECT id, timestamp, method, url, path, query_parameters, headers, body, response_status_code, response_body, response_headers, duration_ms, is_replay, parent_id, replay_count, tags, user_notes, created_at FROM requests WHERE id = ?`

	row := s.db.QueryRowContext(ctx, query, id)
	var req models.WebhookRequest
	var qParams, headers, rHeaders, tags string

	err := row.Scan(
		&req.ID, &req.Timestamp, &req.Method, &req.URL, &req.Path,
		&qParams, &headers, &req.Body, &req.ResponseStatusCode,
		&req.ResponseBody, &rHeaders, &req.DurationMs, &req.IsReplay,
		&req.ParentID, &req.ReplayCount, &tags, &req.UserNotes, &req.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	req.QueryParameters = utils.JSONToMap(qParams)
	req.Headers = utils.JSONToMap(headers)
	req.ResponseHeaders = utils.JSONToMap(rHeaders)
	req.Tags = utils.JSONToSlice(tags)
	return &req, nil
}

func (s *Storage) ListRequests(ctx context.Context, filter models.RequestFilter) ([]*models.WebhookRequest, int, error) {
	where := " WHERE 1=1"
	args := []interface{}{}

	if filter.Method != "" {
		where += " AND method = ?"
		args = append(args, filter.Method)
	}
	if filter.StatusCode > 0 {
		where += " AND response_status_code = ?"
		args = append(args, filter.StatusCode)
	}
	if filter.Query != "" {
		where += " AND (url LIKE ? OR body LIKE ? OR headers LIKE ?)"
		pattern := "%" + filter.Query + "%"
		args = append(args, pattern, pattern, pattern)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM requests"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	where += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, limit, filter.Offset)

	query := `SELECT id, timestamp, method, url, path, query_parameters, headers, body, response_status_code, response_body, response_headers, duration_ms, is_replay, parent_id, replay_count, tags, user_notes, created_at FROM requests` + where

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var list []*models.WebhookRequest
	for rows.Next() {
		var req models.WebhookRequest
		var qParams, headers, rHeaders, tags string
		if err := rows.Scan(
			&req.ID, &req.Timestamp, &req.Method, &req.URL, &req.Path,
			&qParams, &headers, &req.Body, &req.ResponseStatusCode,
			&req.ResponseBody, &rHeaders, &req.DurationMs, &req.IsReplay,
			&req.ParentID, &req.ReplayCount, &tags, &req.UserNotes, &req.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		req.QueryParameters = utils.JSONToMap(qParams)
		req.Headers = utils.JSONToMap(headers)
		req.ResponseHeaders = utils.JSONToMap(rHeaders)
		req.Tags = utils.JSONToSlice(tags)
		list = append(list, &req)
	}

	return list, total, nil
}

func (s *Storage) DeleteRequest(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM requests WHERE id = ?", id)
	return err
}

// Replay Repository Implementations
func (s *Storage) SaveReplayLog(ctx context.Context, log *models.ReplayLog) error {
	query := `INSERT INTO replay_logs (id, original_request_id, timestamp, target_url, method, headers_sent, body_sent, response_status_code, response_body, response_headers, duration_ms, error)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		log.ID, log.OriginalRequestID, log.Timestamp, log.TargetURL, log.Method,
		utils.MapToJSON(log.HeadersSent), log.BodySent, log.ResponseStatusCode,
		log.ResponseBody, utils.MapToJSON(log.ResponseHeaders), log.DurationMs, log.Error,
	)
	if err != nil {
		return err
	}
	_, _ = s.db.ExecContext(ctx, "UPDATE requests SET replay_count = replay_count + 1 WHERE id = ?", log.OriginalRequestID)
	return nil
}

func (s *Storage) GetReplayLogs(ctx context.Context, requestID string) ([]*models.ReplayLog, error) {
	query := `SELECT id, original_request_id, timestamp, target_url, method, headers_sent, body_sent, response_status_code, response_body, response_headers, duration_ms, error FROM replay_logs WHERE original_request_id = ? ORDER BY timestamp DESC`

	rows, err := s.db.QueryContext(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var logs []*models.ReplayLog
	for rows.Next() {
		var l models.ReplayLog
		var headers, rHeaders string
		if err := rows.Scan(
			&l.ID, &l.OriginalRequestID, &l.Timestamp, &l.TargetURL, &l.Method,
			&headers, &l.BodySent, &l.ResponseStatusCode, &l.ResponseBody,
			&rHeaders, &l.DurationMs, &l.Error,
		); err != nil {
			return nil, err
		}
		l.HeadersSent = utils.JSONToMap(headers)
		l.ResponseHeaders = utils.JSONToMap(rHeaders)
		logs = append(logs, &l)
	}
	return logs, nil
}
