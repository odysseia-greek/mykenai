package strategos

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/odysseia-greek/agora/plato/logging"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

type Store struct {
	resultsDir string

	mu       sync.RWMutex
	runs     []RunSummary // sorted newest-first
	byID     map[string]RunSummary
	lastScan time.Time
}

func NewStore(resultsDir string) *Store {
	return &Store{
		resultsDir: resultsDir,
		byID:       make(map[string]RunSummary),
	}
}

func (s *Store) StartPolling(interval time.Duration) {
	if err := s.Scan(); err != nil {
		logging.Error(fmt.Sprintf("initial scan failed: %v", err))
	}

	t := time.NewTicker(interval)
	go func() {
		for range t.C {
			if err := s.Scan(); err != nil {
				logging.Error(fmt.Sprintf("scan error: %v", err))
			}
		}
	}()
}

func (s *Store) Scan() error {
	entries, err := os.ReadDir(s.resultsDir)
	if err != nil {
		return err
	}

	type found struct {
		id   string
		path string
	}

	var foundRuns []found
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
			continue
		}
		p := filepath.Join(s.resultsDir, id, "results.json")
		if _, err := os.Stat(p); err == nil {
			foundRuns = append(foundRuns, found{id: id, path: p})
		}
	}

	summaries := make([]RunSummary, 0, len(foundRuns))
	byID := make(map[string]RunSummary, len(foundRuns))

	for _, r := range foundRuns {
		run, err := readNDJSONResult(r.path, r.id)
		if err != nil {
			logging.Error(fmt.Sprintf("bad results.json id=%s: %v", r.id, err))
			continue
		}

		summary := RunSummary{
			ID:         run.ID,
			Timestamp:  run.Timestamp,
			Passed:     run.Passed,
			Failed:     run.Failed,
			Skipped:    run.Skipped,
			DurationMS: run.Duration.Milliseconds(),
		}
		summaries = append(summaries, summary)
		byID[r.id] = summary
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].ID > summaries[j].ID
	})

	s.mu.Lock()
	s.runs = summaries
	s.byID = byID
	s.lastScan = time.Now().UTC()
	s.mu.Unlock()

	return nil
}

func (s *Store) LastNRuns(n int) []RunSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n > len(s.runs) {
		n = len(s.runs)
	}
	out := make([]RunSummary, n)
	copy(out, s.runs[:n])
	return out
}

func (s *Store) HasRun(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.byID[id]
	return ok
}

func (s *Store) ReadRun(id string) (RunResult, error) {
	if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return RunResult{}, errors.New("invalid run id")
	}
	p := filepath.Join(s.resultsDir, id, "results.json")
	return readNDJSONResult(p, id)
}

func readNDJSONResult(path string, runID string) (RunResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return RunResult{}, err
	}
	defer f.Close()

	outputs := map[string][]string{} // per-test output lines (trimmed)
	runOutput := []string{}          // all output lines (terminal)

	var (
		firstTime  time.Time
		pkgElapsed float64
	)

	tests := map[string]*TestSummary{}
	failures := []string{} // later

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for sc.Scan() {
		raw := bytes.TrimSpace(sc.Bytes())
		if len(raw) == 0 {
			continue
		}

		var ev TestEvent
		if err := json.Unmarshal(raw, &ev); err != nil {
			continue
		}

		if firstTime.IsZero() && !ev.Time.IsZero() {
			firstTime = ev.Time
		}

		switch ev.Action {
		case "run":
			if ev.Test == "" {
				continue
			}
			if _, ok := tests[ev.Test]; !ok {
				tests[ev.Test] = &TestSummary{Name: ev.Test}
			}

		case "pass", "fail", "skip":
			// Package-level completion event has empty Test
			if ev.Test == "" {
				if ev.Elapsed > 0 {
					pkgElapsed = ev.Elapsed
				}
				continue
			}

			ts := tests[ev.Test]
			if ts == nil {
				ts = &TestSummary{Name: ev.Test}
				tests[ev.Test] = ts
			}

			switch ev.Action {
			case "pass":
				ts.Status = "passed"
			case "fail":
				ts.Status = "failed"
				// Build a concise message from captured output
				ts.Message = summarizeFailure(outputs[ev.Test])
			case "skip":
				ts.Status = "skipped"
			}
			ts.DurationMS = int64(ev.Elapsed * 1000)

		case "output":
			// 1) Add to runOutput terminal (keep everything)
			if ev.Output != "" {
				// strip ANSI for browser readability
				line := ansiRE.ReplaceAllString(ev.Output, "")
				line = strings.TrimRight(line, "\n")
				runOutput = append(runOutput, line)

				// cap run output to avoid unbounded memory
				if len(runOutput) > 4000 {
					runOutput = runOutput[len(runOutput)-4000:]
				}
			}

			// 2) Also capture per-test output for failure summary (ignore noise)
			if ev.Test == "" {
				continue
			}
			line := ansiRE.ReplaceAllString(ev.Output, "")

			if strings.HasPrefix(line, "=== RUN") ||
				strings.HasPrefix(line, "--- PASS") ||
				strings.HasPrefix(line, "--- FAIL") {
				continue
			}

			outputs[ev.Test] = append(outputs[ev.Test], strings.TrimRight(line, "\n"))
			if len(outputs[ev.Test]) > 50 {
				outputs[ev.Test] = outputs[ev.Test][len(outputs[ev.Test])-50:]
			}

		default:
			continue
		}
	}

	if err := sc.Err(); err != nil {
		return RunResult{}, err
	}

	outTests := make([]TestSummary, 0, len(tests))
	var passed, failed, skipped int
	for _, t := range tests {
		switch t.Status {
		case "passed":
			passed++
		case "failed":
			failed++
		case "skipped":
			skipped++
		}
		outTests = append(outTests, *t)
	}

	sort.Slice(outTests, func(i, j int) bool { return outTests[i].Name < outTests[j].Name })

	rr := RunResult{
		ID:        runID,
		Timestamp: firstTime.UTC().Format(time.RFC3339),
		Passed:    passed,
		Failed:    failed,
		Skipped:   skipped,
		Tests:     outTests,
		Failures:  failures,
		Output:    runOutput, // ✅ terminal output
	}

	if pkgElapsed > 0 {
		rr.Duration = time.Duration(pkgElapsed * float64(time.Second))
	} else {
		var sumMS int64
		for _, t := range outTests {
			sumMS += t.DurationMS
		}
		rr.Duration = time.Duration(sumMS) * time.Millisecond
	}

	return rr, nil
}

func summarizeFailure(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	// Prefer a "file.go:line:" header, and then include following indented lines.
	// Example:
	// "    critical_ns_test.go:65: not all pods ready..."
	// "        - cilium/... (Pending)"
	var (
		headerIdx = -1
	)

	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		// heuristic: looks like "    file.go:123: ..."
		if strings.Contains(l, ".go:") && strings.Contains(l, ":") {
			headerIdx = i
			break
		}
	}

	// If we found a header, take it + up to next 10 lines that look like detail lines.
	if headerIdx != -1 {
		out := []string{strings.TrimSpace(lines[headerIdx])}
		for j := headerIdx + 1; j < len(lines) && len(out) < 8; j++ {
			l := lines[j]
			// Keep indented/bullet/detail lines, drop empties
			if strings.TrimSpace(l) == "" {
				continue
			}
			if strings.HasPrefix(l, "        ") || strings.HasPrefix(strings.TrimSpace(l), "- ") {
				out = append(out, strings.TrimRight(l, "\n"))
				continue
			}
			// stop when we hit a non-detail line
			break
		}
		return strings.Join(out, "\n")
	}

	// Fallback: last few non-empty lines
	out := make([]string, 0, 5)
	for i := len(lines) - 1; i >= 0 && len(out) < 5; i-- {
		l := strings.TrimSpace(lines[i])
		if l == "" {
			continue
		}
		out = append(out, l)
	}
	// reverse
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return strings.Join(out, "\n")
}
