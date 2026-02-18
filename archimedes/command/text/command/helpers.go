package command

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func getFirstPassageURN(client *http.Client, baseURL, versionURN string) (string, error) {
	u := fmt.Sprintf("%s/library/%s/json/", strings.TrimRight(baseURL, "/"), url.PathEscape(versionURN))
	body, _, err := httpGet(client, u)
	if err != nil {
		return "", err
	}

	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return "", err
	}

	fpAny, ok := m["first_passage"]
	if !ok || fpAny == nil {
		return "", fmt.Errorf("first_passage missing in %s", u)
	}

	if fp, ok := fpAny.(string); ok && fp != "" {
		return fp, nil
	}

	if fpObj, ok := fpAny.(map[string]any); ok {
		if urn, _ := fpObj["urn"].(string); urn != "" {
			return urn, nil
		}
	}

	return "", fmt.Errorf("first_passage present but unsupported type (%T) in %s", fpAny, u)
}

func httpGet(client *http.Client, u string) ([]byte, http.Header, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "archimedes-text-crawler/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, resp.Header, fmt.Errorf("GET %s: %s (%s)", u, resp.Status, strings.TrimSpace(string(b)))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return b, resp.Header, nil
}

func parseNextURNFromLink(link string) string {
	if link == "" {
		return ""
	}
	parts := strings.Split(link, ",")
	re := regexp.MustCompile(`rel="next";\s*urn="([^"]+)"`)
	for _, p := range parts {
		m := re.FindStringSubmatch(p)
		if len(m) == 2 {
			return m[1]
		}
	}
	return ""
}

func citationTail(baseURN, fullURN string) string {
	prefix := baseURN + ":"
	return strings.TrimPrefix(fullURN, prefix)
}

func scaifeReaderLink(baseURL, passageURN string) string {
	return fmt.Sprintf("%s/reader/%s/", strings.TrimRight(baseURL, "/"), url.PathEscape(passageURN))
}

func bookFromRef(ref string) string {
	// "1.1" -> "1"; "1.1.4" -> "1"
	if ref == "" {
		return ""
	}
	if i := strings.Index(ref, "."); i != -1 {
		return ref[:i]
	}
	return ref
}

func safeRefForFilename(ref string) string {
	ref = strings.ReplaceAll(ref, "/", "_")
	ref = strings.ReplaceAll(ref, ":", "_")
	return ref
}

func writeChunkFile(outDir string, chunk WorkChunk) error {
	if outDir == "" {
		return fmt.Errorf("--out-dir is required for crawl (writes one file per reference)")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outDir, err)
	}

	fn := filepath.Join(outDir, fmt.Sprintf("rhema-%s.json", safeRefForFilename(chunk.Reference)))
	f, err := os.Create(fn)
	if err != nil {
		return fmt.Errorf("create %s: %w", fn, err)
	}
	defer f.Close()

	out := RhemaFile{chunk}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode %s: %w", fn, err)
	}
	return nil
}

// splitURN("urn:cts:...perseus-grc2:1.1.0-1.1.4") => ("urn:cts:...perseus-grc2", "1.1.0-1.1.4")
func splitURN(full string) (base string, passage string, err error) {
	i := strings.LastIndex(full, ":")
	if i == -1 || i+1 >= len(full) {
		return "", "", fmt.Errorf("invalid URN (missing passage part): %q", full)
	}
	return full[:i], full[i+1:], nil
}

// expandLastNumericRange("1.1.0-1.1.4") => ["1.1.0","1.1.1","1.1.2","1.1.3","1.1.4"], parent="1.1"
func expandLastNumericRange(p string) (refs []string, parent string, ok bool) {
	if !strings.Contains(p, "-") {
		return nil, "", false
	}
	parts := strings.SplitN(p, "-", 2)
	if len(parts) != 2 {
		return nil, "", false
	}
	start := parts[0]
	end := parts[1]

	// Must share the same prefix up to the last "."
	si := strings.LastIndex(start, ".")
	ei := strings.LastIndex(end, ".")
	if si == -1 || ei == -1 {
		return nil, "", false
	}
	startPrefix := start[:si] // "1.1"
	endPrefix := end[:ei]
	if startPrefix != endPrefix {
		return nil, "", false
	}

	// Last segments must be integers
	aStr := start[si+1:]
	bStr := end[ei+1:]
	a, err1 := strconv.Atoi(aStr)
	b, err2 := strconv.Atoi(bStr)
	if err1 != nil || err2 != nil {
		return nil, "", false
	}
	if a > b {
		a, b = b, a
	}

	parent = startPrefix // "1.1"

	refs = make([]string, 0, (b-a)+1)
	for i := a; i <= b; i++ {
		refs = append(refs, fmt.Sprintf("%s.%d", startPrefix, i))
	}
	return refs, parent, true
}

func parentFromPassage(p string) string {
	// Handle ranges like "1.1.0-1.1.4" => parent "1.1"
	if strings.Contains(p, "-") {
		parts := strings.SplitN(p, "-", 2)
		left := parts[0] // "1.1.0"
		if i := strings.LastIndex(left, "."); i != -1 {
			return left[:i] // "1.1"
		}
		return ""
	}

	// Non-range: "1.1.4" -> "1.1"
	if i := strings.LastIndex(p, "."); i != -1 {
		return p[:i]
	}
	return ""
}
func lastSegment(p string) string {
	if i := strings.LastIndex(p, "."); i != -1 && i+1 < len(p) {
		return p[i+1:]
	}
	return p
}

// When you pass a range, we want to map each generated Greek ref to an English ref.
// For simple aligned editions, the *passage ref* is the same. So we just return r.
func mapPassageToEng(grcPassage, engPassage, currentRef string) string {
	// Map the Greek incipit "....0" to English "....pr" when present.
	// Example: 1.1.0 -> 1.1.pr
	if strings.HasSuffix(currentRef, ".0") {
		return strings.TrimSuffix(currentRef, ".0") + ".pr"
	}
	return currentRef
}

func getNextURN(client *http.Client, baseURL, passageURN string) (string, error) {
	jsonURL := fmt.Sprintf("%s/library/passage/%s/json/", strings.TrimRight(baseURL, "/"), url.PathEscape(passageURN))
	_, hdr, err := httpGet(client, jsonURL)
	if err != nil {
		return "", err
	}
	return parseNextURNFromLink(hdr.Get("Link")), nil
}

func isContainerRef(ref string) bool {
	// "1.2" -> container (two levels)
	// "1.2.3" -> leaf (three levels)
	// crude but works for these CTS refs
	return strings.Count(ref, ".") == 1 && !strings.Contains(ref, "-")
}

func expandLeafRefs(
	client *http.Client,
	baseURL string,
	grcBaseURN string,
	sectionFull string,
	maxLeaf int,
	maxMisses int,
) []string {

	// If it's NOT a range:
	if !strings.Contains(sectionFull, "-") {
		// container like "1.2" => probe .0/.1 and walk forward
		if isContainerRef(sectionFull) {
			return expandContainer(client, baseURL, grcBaseURN, sectionFull, maxLeaf, maxMisses)
		}
		return []string{sectionFull}
	}

	// It IS a range:
	start, end, ok := splitRange(sectionFull)
	if !ok {
		return []string{sectionFull}
	}

	// Easy case: same-prefix range like 1.1.0-1.1.4
	if expanded, _, ok := expandLastNumericRange(sectionFull); ok {
		return expanded
	}

	// Hard case: cross-prefix range like 1.2.1-1.3.2
	sp, ok1 := parseRefParts(start)
	ep, ok2 := parseRefParts(end)
	if !ok1 || !ok2 {
		// Can't parse into 3-level numeric refs; fallback to single
		return []string{sectionFull}
	}

	refs := make([]string, 0, 64)
	misses := 0
	steps := 0

	for sp.lessOrEqual(ep) && steps < maxLeaf {
		steps++

		ref := sp.String()
		u := grcBaseURN + ":" + ref

		_, err := fetchPassagePlainFromCTSXML(client, baseURL, u)
		if err == nil {
			refs = append(refs, ref)
			misses = 0
			sp.sec++
			continue
		}

		// missing
		misses++
		if misses < maxMisses {
			sp.sec++
			continue
		}

		// too many consecutive misses: assume we hit end of this chapter.
		// Move to next chapter and reset sec.
		misses = 0
		sp.chapter++
		sp.sec = 0 // will probe 0,1,2,... in new chapter
	}

	if len(refs) == 0 {
		return []string{sectionFull}
	}
	return refs
}

func splitRange(r string) (start, end string, ok bool) {
	parts := strings.SplitN(r, "-", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	start = strings.TrimSpace(parts[0])
	end = strings.TrimSpace(parts[1])
	if start == "" || end == "" {
		return "", "", false
	}
	return start, end, true
}

func expandContainer(
	client *http.Client,
	baseURL string,
	grcBaseURN string,
	container string, // e.g. "1.2"
	maxLeaf int,
	maxMisses int,
) []string {
	// Try to find a start: .0 or .1
	try := func(n int) bool {
		u := grcBaseURN + ":" + fmt.Sprintf("%s.%d", container, n)
		_, err := fetchPassagePlainFromCTSXML(client, baseURL, u)
		return err == nil
	}

	start := -1
	if try(0) {
		start = 0
	} else if try(1) {
		start = 1
	} else {
		return []string{container}
	}

	refs := make([]string, 0, 32)
	misses := 0
	for i := start; i < start+maxLeaf; i++ {
		ref := fmt.Sprintf("%s.%d", container, i)
		u := grcBaseURN + ":" + ref

		_, err := fetchPassagePlainFromCTSXML(client, baseURL, u)
		if err != nil {
			misses++
			if misses >= maxMisses {
				break
			}
			continue
		}
		misses = 0
		refs = append(refs, ref)
	}
	if len(refs) == 0 {
		return []string{container}
	}
	return refs
}

type refParts struct {
	book    int
	chapter int
	sec     int
}

// parses "1.2.3" into ints
func parseRefParts(ref string) (refParts, bool) {
	parts := strings.Split(ref, ".")
	if len(parts) != 3 {
		return refParts{}, false
	}
	b, err1 := strconv.Atoi(parts[0])
	c, err2 := strconv.Atoi(parts[1])
	s, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return refParts{}, false
	}
	return refParts{book: b, chapter: c, sec: s}, true
}

func (p refParts) String() string {
	return fmt.Sprintf("%d.%d.%d", p.book, p.chapter, p.sec)
}

func (p refParts) lessOrEqual(q refParts) bool {
	if p.book != q.book {
		return p.book < q.book
	}
	if p.chapter != q.chapter {
		return p.chapter < q.chapter
	}
	return p.sec <= q.sec
}
