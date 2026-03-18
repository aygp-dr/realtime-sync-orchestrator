package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	if len(m.sources) != 6 {
		t.Fatalf("expected 6 sources, got %d", len(m.sources))
	}
	if m.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.cursor)
	}
	if m.currentView != viewDashboard {
		t.Errorf("initial view = %d, want dashboard", m.currentView)
	}
}

func TestDashboardView(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	v := m.View()

	if !strings.Contains(v, "RealTimeSyncOrchestrator") {
		t.Error("dashboard missing title")
	}
	if !strings.Contains(v, "postgres-primary") {
		t.Error("dashboard missing postgres-primary source")
	}
	if !strings.Contains(v, "kafka-stream") {
		t.Error("dashboard missing kafka-stream source")
	}
	if !strings.Contains(v, "SOURCE") {
		t.Error("dashboard missing header")
	}
}

func TestNavigateDown(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	if m.cursor != 0 {
		t.Fatalf("expected initial cursor 0")
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", m.cursor)
	}
}

func TestNavigateUp(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	// Move down first
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", m.cursor)
	}
}

func TestNavigateBounds(t *testing.T) {
	m := NewModel(42, 5*time.Second)

	// Try going up at 0
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("cursor went below 0: %d", m.cursor)
	}

	// Go to last item
	for range 10 {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = updated.(Model)
	}
	if m.cursor != 5 {
		t.Errorf("cursor should be capped at 5, got %d", m.cursor)
	}
}

func TestEnterDetailView(t *testing.T) {
	m := NewModel(42, 5*time.Second)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.currentView != viewDetail {
		t.Errorf("view = %d, want detail", m.currentView)
	}

	v := m.View()
	if !strings.Contains(v, "postgres-primary") {
		t.Error("detail view missing source name")
	}
	if !strings.Contains(v, "Sync History") {
		t.Error("detail view missing history section")
	}
}

func TestEscapeFromDetail(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = updated.(Model)
	if m.currentView != viewDashboard {
		t.Errorf("view after esc = %d, want dashboard", m.currentView)
	}
}

func TestQFromDetailGoesBack(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(Model)
	if m.currentView != viewDashboard {
		t.Errorf("view after q in detail = %d, want dashboard", m.currentView)
	}
}

func TestHelpView(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)

	if m.currentView != viewHelp {
		t.Fatalf("view = %d, want help", m.currentView)
	}

	v := m.View()
	if !strings.Contains(v, "Help") {
		t.Error("help view missing title")
	}
	if !strings.Contains(v, "Status Legend") {
		t.Error("help view missing status legend")
	}
}

func TestHelpToggle(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	// Open help
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	// Close help
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	if m.currentView != viewDashboard {
		t.Errorf("view after toggling help = %d, want dashboard", m.currentView)
	}
}

func TestTickRefreshes(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	origLag := m.sources[0].LagMs

	// Multiple ticks to ensure at least one change (random can repeat)
	changed := false
	for range 10 {
		updated, _ := m.Update(tickMsg(time.Now()))
		m = updated.(Model)
		if m.sources[0].LagMs != origLag {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("sources did not refresh after tick messages")
	}
}

func TestSourcesAccessor(t *testing.T) {
	m := NewModel(42, 5*time.Second)
	sources := m.Sources()
	if len(sources) != 6 {
		t.Errorf("Sources() returned %d items, want 6", len(sources))
	}
}

func TestFormatElapsed(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "just now"},
		{5 * time.Second, "5s ago"},
		{90 * time.Second, "1m30s ago"},
	}
	for _, tt := range tests {
		got := formatElapsed(tt.d)
		if got != tt.want {
			t.Errorf("formatElapsed(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
