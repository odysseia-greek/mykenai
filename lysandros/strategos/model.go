package strategos

import "time"

type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`            // run, pass, fail, skip, output, start
	Package string    `json:"Package,omitempty"` // present on most lines
	Test    string    `json:"Test,omitempty"`    // empty for package-level events
	Elapsed float64   `json:"Elapsed,omitempty"` // seconds (usually on pass/fail/skip)
	Output  string    `json:"Output,omitempty"`  // ANSI output etc.
}

type RunResult struct {
	ID        string
	Timestamp string
	Duration  time.Duration

	Passed  int
	Failed  int
	Skipped int

	Tests    []TestSummary
	Failures []string // reserved for later (when you parse output)
	Output   []string
}

type TestSummary struct {
	Name       string
	Status     string // passed|failed|skipped
	DurationMS int64
	Message    string // reserved for later (from output)
}

type RunSummary struct {
	ID         string
	Timestamp  string
	Passed     int
	Failed     int
	Skipped    int
	DurationMS int64
}
