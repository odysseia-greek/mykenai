package command

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/spf13/cobra"
)

func Crawl() *cobra.Command {
	var (
		baseURL     string
		grcBaseURN  string
		engBaseURN  string
		author      string
		book        string
		outDir      string
		maxPassages int
		maxLeaf     int
		maxMisses   int
		timeout     time.Duration
	)

	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Crawl Scaife passages and emit one rhema-<ref>.json per section group (e.g. 1.1, 1.2)",
		RunE: func(cmd *cobra.Command, args []string) error {
			logging.Info("starting text crawl")
			if grcBaseURN == "" || engBaseURN == "" {
				return errors.New("--grc and --eng are required (URN bases)")
			}
			if author == "" || book == "" {
				return errors.New("--author and --book are required")
			}
			if baseURL == "" {
				baseURL = "https://scaife.perseus.org"
			}
			if outDir == "" {
				return errors.New("--out-dir is required (writes one file per reference)")
			}
			logging.Debug(fmt.Sprintf("crawl config base=%s author=%s book=%s outDir=%s max=%d", baseURL, author, book, outDir, maxPassages))

			client := &http.Client{Timeout: timeout}

			firstPassageURN, err := getFirstPassageURN(client, baseURL, grcBaseURN)
			if err != nil {
				return fmt.Errorf("get first passage: %w", err)
			}
			logging.Info(fmt.Sprintf("first passage detected: %s", firstPassageURN))

			cur := firstPassageURN
			count := 0
			chunkCount := 0
			rhemaiCount := 0

			var currentParent string
			var currentBook string
			var chunk WorkChunk

			flush := func() error {
				if chunk.Reference == "" || len(chunk.Rhemai) == 0 {
					return nil
				}
				logging.Info(fmt.Sprintf("writing chunk reference=%s rhemai=%d", chunk.Reference, len(chunk.Rhemai)))
				chunkCount++
				rhemaiCount += len(chunk.Rhemai)
				return writeChunkFile(outDir, chunk)
			}

			for cur != "" {
				if maxPassages > 0 && count >= maxPassages {
					logging.Info(fmt.Sprintf("reached max passages limit (%d), stopping crawl", maxPassages))
					break
				}
				count++

				sectionFull := citationTail(grcBaseURN, cur) // e.g. "1.1.0-1.1.4" OR "1.1.3"

				// Determine parent (file name/group)
				parent := parentFromPassage(sectionFull)
				if parent == "" {
					logging.Debug(fmt.Sprintf("could not resolve parent from section=%s, stopping", sectionFull))
					break
				}

				// Enforce one book
				if currentBook == "" {
					currentBook = bookFromRef(parent)
				} else if bookFromRef(parent) != currentBook {
					logging.Info(fmt.Sprintf("book changed from %s to %s, stopping", currentBook, bookFromRef(parent)))
					break
				}

				// If parent changed, flush and start new file
				if currentParent == "" {
					currentParent = parent
					chunk = WorkChunk{
						Author:          author,
						Book:            book,
						Type:            "work",
						Reference:       currentParent,
						PerseusTextLink: scaifeReaderLink(baseURL, grcBaseURN+":"+currentParent),
						Rhemai:          []Rhema{},
					}
				} else if parent != currentParent {
					if err := flush(); err != nil {
						return err
					}
					currentParent = parent
					chunk = WorkChunk{
						Author:          author,
						Book:            book,
						Type:            "work",
						Reference:       currentParent,
						PerseusTextLink: scaifeReaderLink(baseURL, grcBaseURN+":"+currentParent),
						Rhemai:          []Rhema{},
					}
				}

				// Compute next URN once per crawl-unit (range or single)
				nextURN, err := getNextURN(client, baseURL, cur)
				if err != nil {
					return fmt.Errorf("get next for %s: %w", cur, err)
				}
				logging.Debug(fmt.Sprintf("processing passage=%s section=%s next=%s", cur, sectionFull, nextURN))

				refs := expandLeafRefs(client, baseURL, grcBaseURN, sectionFull, maxLeaf, maxMisses)
				logging.Debug(fmt.Sprintf("expanded %d leaf refs for section=%s", len(refs), sectionFull))

				for _, rref := range refs {
					refParent := parentFromPassage(rref) // "1.2" for "1.2.3", "1.3" for "1.3.1"
					if refParent == "" {
						// fallback: keep using currentParent
						refParent = currentParent
					}

					// If this ref belongs to a different parent (e.g. 1.3), flush and start a new file.
					if currentParent == "" || refParent != currentParent {
						if err := flush(); err != nil {
							return err
						}
						currentParent = refParent
						chunk = WorkChunk{
							Author:          author,
							Book:            book,
							Type:            "work",
							Reference:       currentParent,
							PerseusTextLink: scaifeReaderLink(baseURL, grcBaseURN+":"+currentParent),
							Rhemai:          []Rhema{},
						}
					}

					grcURN := grcBaseURN + ":" + rref
					logging.Debug(fmt.Sprintf("fetching greek urn=%s", grcURN))
					greek, err := fetchPassagePlainFromCTSXML(client, baseURL, grcURN)
					if err != nil {
						return fmt.Errorf("fetch greek (ctsxml) %s: %w", grcURN, err)
					}

					engRef := mapPassageToEng("", "", rref) // 1.1.0 -> 1.1.pr
					engURN := engBaseURN + ":" + engRef

					translation, err := fetchPassagePlainFromCTSXML(client, baseURL, engURN)
					if err != nil {
						logging.Debug(fmt.Sprintf("no english translation found for urn=%s", engURN))
						translation = ""
					}

					greek = normalizeWhitespace(greek)
					translation = normalizeWhitespace(translation)

					rr := Rhema{
						Greek:   greek,
						Section: lastSegment(rref),
					}
					if translation != "" {
						rr.Translations = []string{translation}
					} else {
						rr.Translations = nil // keep nulls
					}

					chunk.Rhemai = append(chunk.Rhemai, rr)
				}

				cur = nextURN
			}

			// Flush last chunk
			if err := flush(); err != nil {
				return err
			}
			logging.Info(fmt.Sprintf("crawl finished passages=%d chunks=%d rhemai=%d", count, chunkCount, rhemaiCount))
			return nil
		},
	}

	cmd.Flags().StringVar(&baseURL, "base", "https://scaife.perseus.org", "Scaife base URL")
	cmd.Flags().StringVar(&grcBaseURN, "grc", "", "Greek version URN base (e.g. urn:cts:...perseus-grc2)")
	cmd.Flags().StringVar(&engBaseURN, "eng", "", "English version URN base (e.g. urn:cts:...perseus-eng2)")
	cmd.Flags().StringVar(&author, "author", "", "Author label for output JSON")
	cmd.Flags().StringVar(&book, "book", "", "Book label for output JSON")
	cmd.Flags().StringVar(&outDir, "out-dir", "", "Directory to write rhema-<ref>.json files (one per section group)")
	cmd.Flags().IntVar(&maxPassages, "max", 0, "Safety limit (0 = no limit)")
	cmd.Flags().IntVar(&maxLeaf, "max-leaf", 100, "Max leaf sections to try per container ref (e.g. 1.2 -> 1.2.1..)")
	cmd.Flags().IntVar(&maxMisses, "max-misses", 2, "Stop leaf probing after N consecutive missing sections")
	cmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP timeout")

	return cmd
}
