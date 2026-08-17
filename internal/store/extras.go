// Copyright 2026 and contributors. Licensed under Apache-2.0. See LICENSE.

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// migrateExtras runs after the generated store migrations and before the
// schema-version stamp. It is the canonical place for novel-feature auxiliary
// tables that need to live in the local store.
//
// Edit this file when adding tables for novel commands. Keep migrations
// idempotent with CREATE TABLE IF NOT EXISTS / CREATE INDEX IF NOT EXISTS so
// every store open can safely re-run them.
func (s *Store) migrateExtras(ctx context.Context, conn *sql.Conn) error {
	migrations := []string{
		// Add CREATE TABLE IF NOT EXISTS statements here.
	}
	for _, m := range migrations {
		if _, err := conn.ExecContext(ctx, m); err != nil {
			return fmt.Errorf("extra migration failed: %w", err)
		}
	}
	return nil
}

// Domain upserts keep timetable resources on typed store methods instead of
// only the generic resources.UpsertBatch path.
func (s *Store) UpsertStation(items []json.RawMessage) (int, int, error) {
	return s.UpsertBatch("station", items)
}

func (s *Store) UpsertFchg(items []json.RawMessage) (int, int, error) {
	return s.UpsertBatch("fchg", items)
}

func (s *Store) UpsertRchg(items []json.RawMessage) (int, int, error) {
	return s.UpsertBatch("rchg", items)
}

func (s *Store) UpsertPlan(items []json.RawMessage) (int, int, error) {
	return s.UpsertBatch("plan", items)
}

func (s *Store) UpsertTimetableResource(resourceType string, items []json.RawMessage) (int, int, error) {
	switch resourceType {
	case "station":
		return s.UpsertStation(items)
	case "fchg":
		return s.UpsertFchg(items)
	case "rchg":
		return s.UpsertRchg(items)
	case "plan":
		return s.UpsertPlan(items)
	default:
		return s.UpsertBatch(resourceType, items)
	}
}

// ResourceCount is a GROUP BY row for local timetable coverage.
type ResourceCount struct {
	ResourceType string `json:"resource_type"`
	Count        int    `json:"count"`
}

func (s *Store) SearchStation(query string, limit int) ([]json.RawMessage, error) {
	return s.Search(query, limit, "station")
}

func (s *Store) SearchFchg(query string, limit int) ([]json.RawMessage, error) {
	return s.Search(query, limit, "fchg")
}

func (s *Store) SearchRchg(query string, limit int) ([]json.RawMessage, error) {
	return s.Search(query, limit, "rchg")
}

func (s *Store) SearchPlan(query string, limit int) ([]json.RawMessage, error) {
	return s.Search(query, limit, "plan")
}

func (s *Store) CountByResourceType() ([]ResourceCount, error) {
	rows, err := s.db.Query(`SELECT resource_type, COUNT(*) FROM resources GROUP BY resource_type ORDER BY resource_type`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ResourceCount
	for rows.Next() {
		var row ResourceCount
		if err := rows.Scan(&row.ResourceType, &row.Count); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

