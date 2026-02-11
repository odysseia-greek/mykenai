package strategos

import (
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

type LysandrosHandler struct {
	store     *Store
	tmplIndex *template.Template
	tmplRun   *template.Template
}

func (l *LysandrosHandler) handleRun(w http.ResponseWriter, r *http.Request, id string) {
	run, err := l.store.ReadRun(id)
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
	if err := l.tmplRun.Execute(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (l *LysandrosHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/")
	if p == "lysandros" || p == "lysandros/" {
		l.handleIndex(w, r)
		return
	}

	id := path.Base(r.URL.Path)
	l.handleRun(w, r, id)
}

func (l *LysandrosHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	runs := l.store.LastNRuns(10)

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
	_ = l.tmplIndex.Execute(w, v)
}
