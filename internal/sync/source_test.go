package sync

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"
)

func TestStatusFromLag(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		lagMs    int64
		lastSync time.Time
		want     Status
	}{
		{"synced at zero", 0, now, StatusSynced},
		{"synced at boundary", 100, now, StatusSynced},
		{"lagging", 500, now, StatusLagging},
		{"lagging upper", 5000, now, StatusLagging},
		{"stale by lag", 5001, now, StatusStale},
		{"stale by time", 50, now.Add(-120 * time.Second), StatusStale},
		{"error negative", -1, now, StatusError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusFromLag(tt.lagMs, tt.lastSync)
			if got != tt.want {
				t.Errorf("StatusFromLag(%d, ...) = %q, want %q", tt.lagMs, got, tt.want)
			}
		})
	}
}

func TestMockSources(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	sources := MockSources(rng)

	if len(sources) != 6 {
		t.Fatalf("expected 6 sources, got %d", len(sources))
	}

	names := map[string]bool{}
	for _, s := range sources {
		names[s.Name] = true
		if s.Target == "" {
			t.Errorf("source %q has empty target", s.Name)
		}
		if s.Name == "" {
			t.Error("source has empty name")
		}
		if len(s.History) != 10 {
			t.Errorf("source %q has %d history entries, want 10", s.Name, len(s.History))
		}
	}

	expectedNames := []string{
		"postgres-primary", "postgres-replica", "redis-cache",
		"elasticsearch", "s3-backup", "kafka-stream",
	}
	for _, n := range expectedNames {
		if !names[n] {
			t.Errorf("missing expected source %q", n)
		}
	}
}

func TestRefreshSource(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	sources := MockSources(rng)
	original := sources[0]
	origHistLen := len(original.History)

	refreshed := RefreshSource(rng, original)

	if refreshed.Name != original.Name {
		t.Errorf("name changed: %q -> %q", original.Name, refreshed.Name)
	}
	if refreshed.Target != original.Target {
		t.Errorf("target changed: %q -> %q", original.Target, refreshed.Target)
	}
	if len(refreshed.History) != origHistLen+1 {
		t.Errorf("history length = %d, want %d", len(refreshed.History), origHistLen+1)
	}
}

func TestRefreshSourceHistoryCap(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	s := Source{
		Name:    "test",
		Target:  "target",
		History: make([]HistoryEntry, 50),
	}
	refreshed := RefreshSource(rng, s)
	if len(refreshed.History) > 50 {
		t.Errorf("history exceeded cap: %d", len(refreshed.History))
	}
}

func TestSourcesJSON(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	sources := MockSources(rng)
	jsonStr, err := SourcesJSON(sources)
	if err != nil {
		t.Fatalf("SourcesJSON error: %v", err)
	}

	var parsed []Source
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(parsed) != 6 {
		t.Errorf("parsed %d sources, want 6", len(parsed))
	}
}

func TestStatusValues(t *testing.T) {
	statuses := []Status{StatusSynced, StatusLagging, StatusStale, StatusError}
	for _, s := range statuses {
		if s == "" {
			t.Error("status constant is empty")
		}
	}
}

func TestRandomLagDistribution(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	counts := map[string]int{"synced": 0, "lagging": 0, "stale": 0, "error": 0}
	n := 1000
	for range n {
		lag := randomLag(rng)
		switch {
		case lag < 0:
			counts["error"]++
		case lag <= 100:
			counts["synced"]++
		case lag <= 5000:
			counts["lagging"]++
		default:
			counts["stale"]++
		}
	}
	// Sanity check: each category should appear at least once in 1000 samples
	for cat, c := range counts {
		if c == 0 {
			t.Errorf("category %q never appeared in %d samples", cat, n)
		}
	}
}
