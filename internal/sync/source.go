package sync

import (
	"encoding/json"
	"math/rand"
	"time"
)

// Status represents the synchronization state of a data source.
type Status string

const (
	StatusSynced  Status = "synced"
	StatusLagging Status = "lagging"
	StatusStale   Status = "stale"
	StatusError   Status = "error"
)

// HistoryEntry records a past sync event.
type HistoryEntry struct {
	Timestamp time.Time     `json:"timestamp"`
	LagMs     int64         `json:"lag_ms"`
	Status    Status        `json:"status"`
	Duration  time.Duration `json:"duration_ns"`
}

// Source represents a data source being synchronized.
type Source struct {
	Name          string         `json:"name"`
	Target        string         `json:"target"`
	LagMs         int64          `json:"lag_ms"`
	LastSync      time.Time      `json:"last_sync"`
	Status        Status         `json:"status"`
	RecordsBehind int64          `json:"records_behind"`
	History       []HistoryEntry `json:"history"`
}

// StatusFromLag derives a Status from a lag value in milliseconds.
func StatusFromLag(lagMs int64, lastSync time.Time) Status {
	if lagMs < 0 {
		return StatusError
	}
	staleCutoff := time.Now().Add(-60 * time.Second)
	if lastSync.Before(staleCutoff) {
		return StatusStale
	}
	if lagMs <= 100 {
		return StatusSynced
	}
	if lagMs <= 5000 {
		return StatusLagging
	}
	return StatusStale
}

// MockSources returns the 6 mock data sources with randomized state.
func MockSources(rng *rand.Rand) []Source {
	now := time.Now()
	defs := []struct {
		name   string
		target string
	}{
		{"postgres-primary", "postgres-replica"},
		{"postgres-replica", "elasticsearch"},
		{"redis-cache", "postgres-primary"},
		{"elasticsearch", "s3-backup"},
		{"s3-backup", "glacier-archive"},
		{"kafka-stream", "elasticsearch"},
	}

	sources := make([]Source, len(defs))
	for i, d := range defs {
		lagMs := randomLag(rng)
		lastSync := now.Add(-time.Duration(rng.Intn(90)) * time.Second)
		status := StatusFromLag(lagMs, lastSync)
		recordsBehind := int64(0)
		if status != StatusSynced {
			recordsBehind = int64(lagMs) * (1 + rng.Int63n(10))
		}

		history := generateHistory(rng, now, 10)

		sources[i] = Source{
			Name:          d.name,
			Target:        d.target,
			LagMs:         lagMs,
			LastSync:      lastSync,
			Status:        status,
			RecordsBehind: recordsBehind,
			History:       history,
		}
	}
	return sources
}

// RefreshSource simulates a data refresh for a single source.
func RefreshSource(rng *rand.Rand, s Source) Source {
	now := time.Now()
	s.LagMs = randomLag(rng)
	s.LastSync = now.Add(-time.Duration(rng.Intn(30)) * time.Second)
	s.Status = StatusFromLag(s.LagMs, s.LastSync)
	if s.Status == StatusSynced {
		s.RecordsBehind = 0
	} else {
		s.RecordsBehind = int64(s.LagMs) * (1 + rng.Int63n(10))
	}

	entry := HistoryEntry{
		Timestamp: now,
		LagMs:     s.LagMs,
		Status:    s.Status,
		Duration:  time.Duration(rng.Intn(500)) * time.Millisecond,
	}
	s.History = append(s.History, entry)
	if len(s.History) > 50 {
		s.History = s.History[len(s.History)-50:]
	}
	return s
}

// SourcesJSON returns the JSON representation of sources.
func SourcesJSON(sources []Source) (string, error) {
	data, err := json.MarshalIndent(sources, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func randomLag(rng *rand.Rand) int64 {
	r := rng.Float64()
	switch {
	case r < 0.4:
		return rng.Int63n(100) // synced: 0-99ms
	case r < 0.7:
		return 100 + rng.Int63n(4900) // lagging: 100-4999ms
	case r < 0.9:
		return 5000 + rng.Int63n(25000) // stale: 5000-29999ms
	default:
		return -1 // error
	}
}

func generateHistory(rng *rand.Rand, now time.Time, count int) []HistoryEntry {
	entries := make([]HistoryEntry, count)
	for i := range entries {
		t := now.Add(-time.Duration(count-i) * 5 * time.Second)
		lag := randomLag(rng)
		entries[i] = HistoryEntry{
			Timestamp: t,
			LagMs:     lag,
			Status:    StatusFromLag(lag, t),
			Duration:  time.Duration(rng.Intn(500)) * time.Millisecond,
		}
	}
	return entries
}
