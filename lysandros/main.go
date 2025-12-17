package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/odysseia-greek/agora/plato/logging"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

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
		log.Printf("[lysandros] initial scan error: %v", err)
	}

	t := time.NewTicker(interval)
	go func() {
		for range t.C {
			if err := s.Scan(); err != nil {
				log.Printf("[lysandros] scan error: %v", err)
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
			log.Printf("[lysandros] bad results.json id=%s: %v", r.id, err)
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

type Server struct {
	store     *Store
	tmplIndex *template.Template
	tmplRun   *template.Template
}

func main() {
	//https://patorjk.com/software/taag/#p=display&f=Crawford2&t=LYSANDROS&x=none&v=4&h=4&w=80&we=false
	logging.System(`
 _      __ __  _____  ____  ____   ___    ____   ___   _____
| |    |  |  |/ ___/ /    ||    \ |   \  |    \ /   \ / ___/
| |    |  |  (   \_ |  o  ||  _  ||    \ |  D  )     (   \_ 
| |___ |  ~  |\__  ||     ||  |  ||  D  ||    /|  O  |\__  |
|     ||___, |/  \ ||  _  ||  |  ||     ||    \|     |/  \ |
|     ||     |\    ||  |  ||  |  ||     ||  .  \     |\    |
|_____||____/  \___||__|__||__|__||_____||__|\_|\___/  \___|
`)
	logging.System("\"λέγεται δὲ ὁ Λυσάνδρου πατὴρ Ἀριστόκλειτος οἰκίας μὲν οὐ γενέσθαι βασιλικῆς, ἄλλως δὲ γένους εἶναι τοῦ τῶν Ἡρακλειδῶν\"")
	logging.System("\"The father of Lysander, Aristocleitus, is said to have been of the lineage of the Heracleidae, though not of the royal family.\"")
	logging.System("starting html viewer.....")

	logging.System("getting env variables and creating config")

	resultsDir := getenv("LYSANDROS_RESULTS_DIR", "./results")
	addr := getenv("LYSANDROS_ADDR", ":8090")

	store := NewStore(resultsDir)
	store.StartPolling(1 * time.Minute)

	tmplIndex := template.Must(template.ParseFS(templatesFS, "templates/index.html"))
	tmplRun := template.Must(template.ParseFS(templatesFS, "templates/run.html"))

	s := &Server{
		store:     store,
		tmplIndex: tmplIndex,
		tmplRun:   tmplRun,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	staticSub := mustSub(staticFS, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	log.Printf("[lysandros] listening on %s, resultsDir=%s", addr, resultsDir)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		s.handleIndex(w, r)
		return
	}
	path = strings.TrimSuffix(path, "/")
	s.handleRun(w, r, path)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	runs := s.store.LastNRuns(10)

	total := len(runs)
	passedRuns := 0
	failedRuns := 0
	totalPassedTests := 0
	totalFailedTests := 0

	streak := 0
	for i, r := range runs {
		if r.Failed == 0 {
			passedRuns++
			if i == 0 || streak >= 0 { // continuing pass streak from newest
				streak++
			}
		} else {
			failedRuns++
			if i == 0 {
				streak = 0
			}
			if i == 0 { /* newest is fail, pass streak stays 0 */
			}
		}
		totalPassedTests += r.Passed
		totalFailedTests += r.Failed
	}

	lastStatus := "—"
	lastID := ""
	if total > 0 {
		lastID = runs[0].ID
		if runs[0].Failed == 0 {
			lastStatus = "PASS"
		} else {
			lastStatus = "FAIL"
		}
	}

	passRate := 0
	if total > 0 {
		passRate = int(float64(passedRuns) / float64(total) * 100.0)
	}

	type view struct {
		Runs []RunSummary
		Now  string

		TotalRunsShown int
		PassedRuns     int
		FailedRuns     int
		PassRatePct    int

		TotalPassedTests int
		TotalFailedTests int

		LatestID     string
		LatestStatus string
		PassStreak   int
	}

	v := view{
		Runs: runs,
		Now:  time.Now().UTC().Format(time.RFC3339),

		TotalRunsShown: total,
		PassedRuns:     passedRuns,
		FailedRuns:     failedRuns,
		PassRatePct:    passRate,

		TotalPassedTests: totalPassedTests,
		TotalFailedTests: totalFailedTests,

		LatestID:     lastID,
		LatestStatus: lastStatus,
		PassStreak:   streak,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.tmplIndex.Execute(w, v)
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request, id string) {
	run, err := s.store.ReadRun(id)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || strings.Contains(err.Error(), "invalid run id") {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type view struct {
		ID  string
		Run RunResult
	}
	v := view{ID: id, Run: run}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmplRun.Execute(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getenv(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}
