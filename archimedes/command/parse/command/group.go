package command

import (
	"encoding/json"
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/agora/plato/models"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

func GroupChapters() *cobra.Command {
	var (
		sullegoPath string
	)
	cmd := &cobra.Command{
		Use:   "group",
		Short: "group a list of words",
		Long: `Allows you to parse a list of words to be used by sokrates
- Filepath
`,
		Run: func(cmd *cobra.Command, args []string) {
			err := createChaptersFromParsed(sullegoPath)
			if err != nil {
				logging.Error(err.Error())
			}

		},
	}
	cmd.PersistentFlags().StringVarP(&sullegoPath, "sullego", "s", "", "where to the sullego directory to parse")

	return cmd
}

func createChaptersFromParsed(sullegoPath string) error {
	// Walk through subdirectories of sullegoPath
	err := filepath.Walk(sullegoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a directory and not in the excludeDirs list
		if info.IsDir() {
			// Check if the directory contains a "logos.json" file
			logosJSONPath := filepath.Join(path, "logos.json")
			if strings.Contains(logosJSONPath, "multichoice") {
				_, err := os.Stat(logosJSONPath)
				if err == nil {
					// Read and parse "logos.json"
					plan, err := os.ReadFile(logosJSONPath)
					if err != nil {
						return err
					}

					var logos models.Logos
					err = json.Unmarshal(plan, &logos)
					if err != nil {
						return err
					}

					parsedGroup, name := groupChapters(logos, path)

					// Determine the output file path
					fileOut := filepath.Join(sullegoPath, "multichoice", name)

					// Write the parsed data to the output file
					err = util.WriteJSONToFilePrettyPrint(parsedGroup, fileOut)
					if err != nil {
						return err
					}

					// Print a message indicating the processing is done
					logging.Info(fmt.Sprintf("Processed and saved data in %s\n", fileOut))
				}
			}
		}

		return nil
	})

	return err
}

func groupChapters(logoi models.Logos, path string) ([]MultiChoiceQuiz, string) {
	var aggrModel []MultiChoiceQuiz

	splitPath := strings.Split(path, "/")
	author := capitalizeFirst(splitPath[len(splitPath)-2])
	book := capitalizeFirst(splitPath[len(splitPath)-1])
	theme := fmt.Sprintf("%s - %s", author, book)
	name := fmt.Sprintf("%s%s.json", author, book)
	var chapter int64
	var content []Content

	for i, word := range logoi.Logos {
		if chapter != word.Chapter && chapter != 0 {
			mcModel := MultiChoiceQuiz{
				QuizType: "MutliChoice",
				Theme:    theme,
				Set:      int(chapter),
				Content:  content,
			}

			mcModel.QuizMetadata.Language = "English"
			aggrModel = append(aggrModel, mcModel)
			content = []Content{}
		}

		content = append(content, Content{
			Translation: word.Translation,
			Greek:       word.Greek,
		})
		chapter = word.Chapter

		if len(logoi.Logos)-1 == i {
			if len(content) > 9 {
				mcModel := MultiChoiceQuiz{
					QuizType: "MutliChoice",
					Theme:    theme,
					Set:      int(chapter),
					Content:  content,
				}

				mcModel.QuizMetadata.Language = "English"
				aggrModel = append(aggrModel, mcModel)
			} else {
				for _, extraContent := range content {
					aggrModel[len(aggrModel)-1].Content = append(aggrModel[len(aggrModel)-1].Content, extraContent)
				}
			}
		}
	}

	return aggrModel, name
}

func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	firstRune, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(firstRune)) + s[size:]
}

type MultiChoiceQuiz struct {
	QuizMetadata struct {
		Language string `json:"language"`
	} `json:"quizMetadata"`
	QuizType string    `json:"quizType"`
	Theme    string    `json:"theme,omitempty"`
	Set      int       `json:"set,omitempty"`
	Content  []Content `json:"content"`
	Progress struct {
		TimesCorrect    int     `json:"timesCorrect"`
		TimesIncorrect  int     `json:"timesIncorrect"`
		AverageAccuracy float64 `json:"averageAccuracy"`
	} `json:"progress,omitempty"`
}

type Content struct {
	Translation string `json:"translation"`
	Greek       string `json:"greek,omitempty"`
}
