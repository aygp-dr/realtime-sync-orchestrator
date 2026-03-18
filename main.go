package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aygp-dr/realtime-sync-orchestrator/internal/sync"
	"github.com/aygp-dr/realtime-sync-orchestrator/internal/tui"
)

func main() {
	jsonFlag := flag.Bool("json", false, "Output current sync status as JSON and exit")
	tickFlag := flag.Duration("tick", 5*time.Second, "Dashboard refresh interval")
	flag.Parse()

	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	if *jsonFlag {
		sources := sync.MockSources(rng)
		out, err := sync.SourcesJSON(sources)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(out)
		return
	}

	m := tui.NewModel(seed, *tickFlag)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
