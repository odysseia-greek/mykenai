package command

import (
	"encoding/json"
	"github.com/kpango/glg"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/odysseia-greek/plato/models"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func ReparseList() *cobra.Command {
	var (
		sullegoPath string
		all         bool
	)
	cmd := &cobra.Command{
		Use:   "reparse",
		Short: "reparse a list of words",
		Long: `Allows you to parse a list of words to be used by sokrates
- Filepath
`,
		Run: func(cmd *cobra.Command, args []string) {
			glg.Green("parsing")
			homeDir, _ := os.UserHomeDir()

			var parsedPath string
			if all {
				parsedPath = filepath.Join(homeDir, GODIR, IONIADIR)
			} else {
				parsedPath = filepath.Join(homeDir, GODIR, IONIADIR, sullegoPath)
			}

			parseVerbNounMisc(parsedPath, all)

		},
	}
	cmd.PersistentFlags().StringVarP(&sullegoPath, "sullego", "s", "", "where to the sullego directory to parse")
	cmd.PersistentFlags().BoolVarP(&all, "all", "a", false, "do all the contents of sullego")

	return cmd
}

func parseVerbNounMisc(sullegoPath string, all bool) {
	if all {
		// Walk through subdirectories of sullegoPath
		err := filepath.Walk(sullegoPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if it's a directory and not in the excludeDirs list
			if info.IsDir() {
				// Check if the directory contains a "logos.json" file
				logosJSONPath := filepath.Join(path, "logos.json")
				_, err := os.Stat(logosJSONPath)
				if err == nil {
					// Read and parse "logos.json"
					plan, err := ioutil.ReadFile(logosJSONPath)
					if err != nil {
						return err
					}

					var logos models.Logos
					err = json.Unmarshal(plan, &logos)
					if err != nil {
						return err
					}

					// Process the logos data (you can modify this part)
					parsedLogos := parseOutDifferentPaths(logos)

					// Determine the output file path
					fileOut := filepath.Join(path, "logos.json")

					// Write the parsed data to the output file
					err = util.WriteJSONToFilePrettyPrint(parsedLogos, fileOut)
					if err != nil {
						return err
					}

					// Print a message indicating the processing is done
					glg.Infof("Processed and saved data in %s\n", fileOut)
				}
			}

			return nil
		})

		if err != nil {
			glg.Error(err)
		}
	} else {
		// Process the "logos.json" file in sullegoPath (as in your original code)
		readOut := filepath.Join(sullegoPath, "logos.json")
		plan, err := ioutil.ReadFile(readOut)
		if err != nil {
			glg.Error(err)
		}

		var logos models.Logos
		err = json.Unmarshal(plan, &logos)
		if err != nil {
			glg.Error(err)
		}

		parsedLogos := parseOutDifferentPaths(logos)
		fileOut := filepath.Join(sullegoPath, "logos.json")
		err = util.WriteJSONToFilePrettyPrint(parsedLogos, fileOut)
		if err != nil {
			glg.Error(err)
		}
	}
}

func parseOutDifferentPaths(logoi models.Logos) models.Logos {
	var nounLogos models.Logos
	var namesLogos models.Logos
	var verbLogos models.Logos
	var miscLogos models.Logos

	for _, logos := range logoi.Logos {
		trimmed := strings.TrimSpace(logos.Greek)
		greek := util.RemoveAccent(trimmed)
		// Convert Greek text to lowercase for consistent comparisons
		greekLower := strings.ToLower(greek)

		if unicode.IsUpper([]rune(greek)[0]) {
			namesLogos.Logos = append(namesLogos.Logos, logos)
		} else if len(greekLower) <= 2 {
			miscLogos.Logos = append(miscLogos.Logos, logos)
		} else {
			// Check if Greek text ends with specific suffixes
			if strings.HasSuffix(greekLower, "ος") ||
				strings.HasSuffix(greekLower, "ια") ||
				strings.HasSuffix(greekLower, "η") ||
				strings.HasSuffix(greekLower, "ως") ||
				strings.HasSuffix(greekLower, "ης") ||
				strings.HasSuffix(greekLower, "ις") ||
				strings.HasSuffix(greekLower, "ων") ||
				strings.HasSuffix(greekLower, "ωρ") ||
				strings.HasSuffix(greekLower, "ρα") ||
				strings.HasSuffix(greekLower, "αξ") ||
				strings.HasSuffix(greekLower, "ον") ||
				strings.HasSuffix(greekLower, "ηρ") ||
				strings.HasSuffix(greekLower, "υς") ||
				strings.HasSuffix(greekLower, "μα") {
				nounLogos.Logos = append(nounLogos.Logos, logos)
			} else if strings.HasSuffix(greekLower, "μι") ||
				strings.HasSuffix(greek, "ω") ||
				strings.HasSuffix(greekLower, "μαι") {
				verbLogos.Logos = append(verbLogos.Logos, logos)
			} else {
				miscLogos.Logos = append(miscLogos.Logos, logos)
			}
		}
	}

	nounGroups := groupWordsByLength(nounLogos)
	verbGroups := groupWordsByLength(verbLogos)
	miscGroups := groupWordsByLength(miscLogos)
	namedGroups := groupWordsByLength(namesLogos)

	chapters := createChapters(nounGroups, verbGroups, miscGroups, namedGroups)
	return combineChapters(chapters)
}

func groupWordsByLength(namesLogos models.Logos) map[int][]models.Word {
	wordLengthGroups := make(map[int][]models.Word)

	// Sort the words by length and group them by length
	for _, word := range namesLogos.Logos {
		greek := util.RemoveAccent(strings.TrimSpace(word.Greek))
		wordLength := utf8.RuneCountInString(greek)
		wordLengthGroups[wordLength] = append(wordLengthGroups[wordLength], word)
	}

	return wordLengthGroups
}

func createChapters(groupsList ...map[int][]models.Word) []models.Logos {
	chapters := make([]models.Logos, 0)
	currentChapter := models.Logos{}
	wordCount := 0
	chapterNumber := 1

	// Iterate through word length groups in the order they are provided
	for _, wordLengthGroups := range groupsList {
		// Sort word lengths in ascending order
		var wordLengths []int
		for length := range wordLengthGroups {
			wordLengths = append(wordLengths, length)
		}
		sort.Ints(wordLengths)

		// Iterate through word lengths and add words to chapters
		for _, length := range wordLengths {
			words := wordLengthGroups[length]
			for _, word := range words {
				// Set the Chapter field of each word to the current chapter number
				word.Chapter = int64(chapterNumber)

				// Add the word to the current chapter
				currentChapter.Logos = append(currentChapter.Logos, word)
				wordCount++

				// If the chapter reaches 20 words, start a new chapter
				if wordCount == 20 {
					chapters = append(chapters, currentChapter)
					currentChapter = models.Logos{}
					wordCount = 0
					chapterNumber++
				}
			}
		}
	}

	// Add any remaining words to the last chapter
	if len(currentChapter.Logos) > 0 {
		chapters = append(chapters, currentChapter)
	}

	return chapters
}

func combineChapters(chaptersList ...[]models.Logos) models.Logos {
	combined := models.Logos{Logos: []models.Word{}}
	for _, chapters := range chaptersList {
		for _, chapter := range chapters {
			combined.Logos = append(combined.Logos, chapter.Logos...)
		}
	}
	return combined
}
