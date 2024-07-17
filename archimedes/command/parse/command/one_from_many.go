package command

import (
	"encoding/json"
	"fmt"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func CreateOneFromMany() *cobra.Command {
	var (
		sullegoPath string
	)
	cmd := &cobra.Command{
		Use:   "one-from-many",
		Short: "take all the current files and create new ones",
		Long: `Allows you to parse a list of words to one list used by Sokrates
- Filepath
`,
		Run: func(cmd *cobra.Command, args []string) {
			err := createOneFromMany(sullegoPath)
			if err != nil {
				logging.Error(err.Error())
			}
		},
	}
	cmd.PersistentFlags().StringVarP(&sullegoPath, "sullego", "s", "", "where to the sullego directory to parse")
	return cmd
}

func createOneFromMany(sullegoPath string) error {
	// Create the multichoice directory
	multichoiceDir := filepath.Join(sullegoPath, "multichoice")
	err := os.Mkdir(multichoiceDir, 0777)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Initialize maps to aggregate results
	allParsedGroups := map[string]map[string]Content{
		"nouns":   {},
		"names":   {},
		"verbs":   {},
		"misc":    {},
		"doubles": {},
	}

	// Walk through the authorbased directory
	authorbasedDir := filepath.Join(sullegoPath, "authorbased")
	err = filepath.Walk(authorbasedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a file with .json extension
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			// Read and parse the JSON file
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var multiChoiceQuiz []MultiChoiceQuiz
			err = json.Unmarshal(fileContent, &multiChoiceQuiz)
			if err != nil {
				return err
			}

			for _, quiz := range multiChoiceQuiz {
				// Check if the language is "English"
				if strings.ToLower(quiz.QuizMetadata.Language) == "english" {
					parsedGroups := parseNounVerbMiscName(quiz)

					// Aggregate parsed data into the overall map
					aggregateParsedGroups(allParsedGroups, parsedGroups)
				}
			}
		}
		return nil
	})

	// Write aggregated results to different files
	err = writeAggregatedGroups(allParsedGroups, multichoiceDir)
	if err != nil {
		return err
	}

	return nil
}

func aggregateParsedGroups(aggregatedGroups map[string]map[string]Content, parsedGroups map[string][]Content) {
	for key, contents := range parsedGroups {
		for _, content := range contents {
			greek := strings.TrimSpace(content.Greek)
			if _, exists := aggregatedGroups[key][greek]; exists {
				aggregatedGroups["doubles"][greek] = content
			} else {
				aggregatedGroups[key][greek] = content
			}
		}
	}
}

func writeAggregatedGroups(aggregatedGroups map[string]map[string]Content, outputDir string) error {
	for fileName, dataMap := range aggregatedGroups {
		if fileName == "doubles" {
			data := make([]Content, 0, len(dataMap))
			for _, content := range dataMap {
				data = append(data, content)
			}
			fileOut := filepath.Join(outputDir, fmt.Sprintf("%s.json", fileName))
			err := util.WriteJSONToFilePrettyPrint(data, fileOut)
			if err != nil {
				return err
			}
			continue
		}

		dataList := make([]Content, 0, len(dataMap))
		for _, content := range dataMap {
			dataList = append(dataList, content)
		}

		sort.Slice(dataList, func(i, j int) bool {
			return dataList[i].Greek < dataList[j].Greek
		})

		quizzes := createQuizzesFromContent(fileName, dataList)
		fileOut := filepath.Join(outputDir, fmt.Sprintf("%s.json", fileName))
		err := util.WriteJSONToFilePrettyPrint(quizzes, fileOut)
		if err != nil {
			return err
		}
	}
	return nil
}

func createQuizzesFromContent(theme string, contents []Content) []MultiChoiceQuiz {
	var quizzes []MultiChoiceQuiz
	set := 1
	var currentContent []Content

	for i, content := range contents {
		currentContent = append(currentContent, content)

		if len(currentContent) == 20 || (i == len(contents)-1 && len(currentContent) > 0) {
			if len(currentContent) < 5 && len(quizzes) > 0 {
				quizzes[len(quizzes)-1].Content = append(quizzes[len(quizzes)-1].Content, currentContent...)
			} else {
				quiz := MultiChoiceQuiz{
					QuizMetadata: struct {
						Language string `json:"language"`
					}{
						Language: "English",
					},
					QuizType: "Multiple Choice",
					Theme:    strings.Title(theme),
					Set:      set,
					Content:  currentContent,
				}
				quizzes = append(quizzes, quiz)
				set++
			}
			currentContent = nil
		}
	}

	return quizzes
}

func parseNounVerbMiscName(quiz MultiChoiceQuiz) map[string][]Content {
	var nounLogos, namesLogos, verbLogos, miscLogos []Content

	for _, content := range quiz.Content {
		trimmed := strings.TrimSpace(content.Greek)
		greek := util.RemoveAccent(trimmed)
		greekLower := strings.ToLower(greek)

		if unicode.IsUpper([]rune(greek)[0]) {
			namesLogos = append(namesLogos, content)
		} else if utf8.RuneCountInString(greekLower) <= 2 {
			miscLogos = append(miscLogos, content)
		} else {
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
				strings.HasSuffix(greekLower, "μα") ||
				strings.HasSuffix(greekLower, "ας") ||
				strings.HasSuffix(greekLower, "εα") ||
				strings.HasSuffix(greekLower, "σα") {
				nounLogos = append(nounLogos, content)
			} else if strings.HasSuffix(greekLower, "μι") ||
				strings.HasSuffix(greek, "ω") ||
				strings.HasSuffix(greekLower, "μαι") {
				verbLogos = append(verbLogos, content)
			} else {
				miscLogos = append(miscLogos, content)
			}
		}
	}

	return map[string][]Content{
		"nouns": nounLogos,
		"names": namesLogos,
		"verbs": verbLogos,
		"misc":  miscLogos,
	}
}
