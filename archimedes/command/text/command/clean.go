package command

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func fetchPassagePlainFromCTSXML(client *http.Client, baseURL, urn string) (string, error) {
	u := fmt.Sprintf("%s/library/%s/cts-api-xml/", strings.TrimRight(baseURL, "/"), url.PathEscape(urn))

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "archimedes-text-crawler/0.1")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return "", fmt.Errorf("GET %s: %s (%s)", u, resp.Status, strings.TrimSpace(string(b)))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return extractReadingTextFromCTSXML(bytes.NewReader(b))
}

func extractReadingTextFromCTSXML(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)

	// Skip these tags entirely (subtree)
	skip := map[string]bool{
		"reg":  true,
		"note": true,
	}

	skipDepth := 0
	inPDepth := 0 // >0 means we are inside a <p>...</p>
	var sb strings.Builder

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			// Handle skip subtree first
			if skipDepth > 0 {
				skipDepth++
				continue
			}
			if skip[t.Name.Local] {
				skipDepth = 1
				continue
			}

			// Only capture text inside <p> elements (TEI paragraphs)
			if t.Name.Local == "p" {
				inPDepth++
			} else if inPDepth > 0 {
				// nested element inside <p> - keep inPDepth as-is
			}

		case xml.EndElement:
			if skipDepth > 0 {
				skipDepth--
				continue
			}
			if t.Name.Local == "p" && inPDepth > 0 {
				inPDepth--
				// add a sentence separator between paragraphs
				if sb.Len() > 0 {
					sb.WriteByte(' ')
				}
			}

		case xml.CharData:
			if skipDepth > 0 || inPDepth == 0 {
				continue
			}

			txt := strings.TrimSpace(string(t))
			if txt == "" {
				continue
			}

			if sb.Len() > 0 {
				sb.WriteByte(' ')
			}
			sb.WriteString(txt)
		}
	}

	return normalizePunctuationWhitespace(normalizeWhitespace(sb.String())), nil
}

func normalizeWhitespace(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	return strings.Join(strings.Fields(s), " ")
}

func normalizePunctuationWhitespace(s string) string {
	// Fix spaces before punctuation
	repls := []struct{ from, to string }{
		{" ,", ","},
		{" .", "."},
		{" ;", ";"},
		{" :", ":"},
		{" !", "!"},
		{" ?", "?"},
	}
	for _, r := range repls {
		s = strings.ReplaceAll(s, r.from, r.to)
	}
	return s
}
