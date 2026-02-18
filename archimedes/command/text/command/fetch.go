package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/spf13/cobra"
)

func Fetch() *cobra.Command {
	var (
		baseURL string
		grcURN  string
		engURN  string
		author  string
		book    string
		ref     string
		outPath string
		outDir  string
		timeout time.Duration
	)

	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch a single Scaife passage; if a range is provided, split into individual sections",
		Long: `Fetches a single passage (or range) from Scaife and writes it in your rhema.json shape.

If the URN contains a range like 1.1.0-1.1.4, this command expands it to:
  1.1.0, 1.1.1, 1.1.2, 1.1.3, 1.1.4
		and fetches each passage separately so you get multiple rhemai entries.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logging.Info("starting text fetch")
			if grcURN == "" {
				return errors.New("--grc is required (full Greek passage URN; may include range)")
			}
			if engURN == "" {
				return errors.New("--eng is required (full English passage URN; may include range)")
			}
			if author == "" || book == "" {
				return errors.New("--author and --book are required")
			}
			if baseURL == "" {
				baseURL = "https://scaife.perseus.org"
			}
			logging.Debug(fmt.Sprintf("fetch config base=%s author=%s book=%s out=%s outDir=%s", baseURL, author, book, outPath, outDir))

			client := &http.Client{Timeout: timeout}

			grcBase, grcPassage, err := splitURN(grcURN)
			if err != nil {
				return fmt.Errorf("parse --grc: %w", err)
			}
			engBase, engPassage, err := splitURN(engURN)
			if err != nil {
				return fmt.Errorf("parse --eng: %w", err)
			}

			// Expand range if present, otherwise single passage.
			refs, parentRef, ok := expandLastNumericRange(grcPassage)
			if !ok {
				refs = []string{grcPassage}
				parentRef = parentFromPassage(grcPassage) // best-effort
			}
			logging.Info(fmt.Sprintf("resolved %d passage reference(s), parent=%s", len(refs), parentRef))

			// Default the "reference" field to parentRef if caller didn't provide --ref
			if ref == "" {
				ref = parentRef
			}

			// Top-level link: point to parent passage, like your hand-made files do.
			// (If parentRef is empty, fall back to the original URN.)
			linkURN := grcURN
			if parentRef != "" {
				linkURN = grcBase + ":" + parentRef
			}

			out := RhemaFile{
				{
					Author:          author,
					Book:            book,
					Type:            "work",
					Reference:       ref,
					PerseusTextLink: scaifeReaderLink(baseURL, linkURN),
					Rhemai:          []Rhema{},
				},
			}

			for _, r := range refs {
				grc := grcBase + ":" + r
				eng := engBase + ":" + mapPassageToEng(grcPassage, engPassage, r)
				logging.Debug(fmt.Sprintf("fetching passage greek=%s english=%s", grc, eng))

				greek, err := fetchPassagePlainFromCTSXML(client, baseURL, grc)
				if err != nil {
					return fmt.Errorf("fetch greek %s: %w", grc, err)
				}

				translation, err := fetchPassagePlainFromCTSXML(client, baseURL, eng)
				if err != nil {
					logging.Debug(fmt.Sprintf("no english translation found for urn=%s", eng))
					translation = "" // keep going if translation missing
				}

				greek = normalizeWhitespace(greek)
				translation = normalizeWhitespace(translation)

				section := lastSegment(r)
				item := Rhema{
					Greek:   greek,
					Section: section,
				}
				if translation != "" {
					item.Translations = []string{translation}
				}

				out[0].Rhemai = append(out[0].Rhemai, item)
			}

			if outPath == "" {
				if outDir != "" {
					safeRef := strings.ReplaceAll(ref, "/", "_")
					safeRef = strings.ReplaceAll(safeRef, ":", "_")
					outPath = filepath.Join(outDir, fmt.Sprintf("rhema-%s.json", safeRef))
				} else {
					outPath = "-"
				}
			}
			logging.Info(fmt.Sprintf("writing fetch output to %s", outPath))

			var w io.Writer = os.Stdout
			if outPath != "" && outPath != "-" {
				f, err := os.Create(outPath)
				if err != nil {
					return fmt.Errorf("open output file: %w", err)
				}
				defer f.Close()
				w = f
			}

			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			if err := enc.Encode(out); err != nil {
				return fmt.Errorf("encode json: %w", err)
			}
			logging.Info(fmt.Sprintf("fetch finished with %d rhemai", len(out[0].Rhemai)))
			return nil
		},
	}

	cmd.Flags().StringVar(&baseURL, "base", "https://scaife.perseus.org", "Scaife base URL")
	cmd.Flags().StringVar(&grcURN, "grc", "", "Greek passage URN (full; may include range e.g. ...:1.1.0-1.1.4)")
	cmd.Flags().StringVar(&engURN, "eng", "", "English passage URN (full; may include range e.g. ...:1.1.0-1.1.4)")
	cmd.Flags().StringVar(&author, "author", "", "Author label for output JSON")
	cmd.Flags().StringVar(&book, "book", "", "Book label for output JSON")
	cmd.Flags().StringVar(&ref, "ref", "", "Top-level reference label (defaults to parent passage, e.g. 1.1)")
	cmd.Flags().StringVarP(&outPath, "out", "o", "", "Output path (default stdout; use '-' for stdout)")
	cmd.Flags().StringVar(&outDir, "out-dir", "", "Directory to write rhema-<ref>.json (used if --out not set)")
	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP timeout")

	return cmd
}
