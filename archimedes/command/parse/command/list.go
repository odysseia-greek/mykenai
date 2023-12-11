package command

import (
	"fmt"
	"github.com/odysseia-greek/agora/plato/helpers"
	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/agora/plato/models"
	"github.com/odysseia-greek/mykenai/archimedes/util"
	"github.com/spf13/cobra"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	ByFile   string = "by-file"
	ByLetter string = "by-letter"
)

func ListToWords() *cobra.Command {
	var (
		filePath string
		outDir   string
		mode     string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "parse a list of words",
		Long: `Allows you to parse a list of words to be used by demokritos
- Filepath
- OutDir
`,
		Run: func(cmd *cobra.Command, args []string) {
			if filePath == "" {
				logging.Debug(fmt.Sprintf("filepath is empty"))
				return
			}

			if outDir == "" {
				logging.Debug(fmt.Sprintf("no outdir set assuming one"))
				homeDir, _ := os.UserHomeDir()
				outDir = fmt.Sprintf("%s/go/src/github.com/odysseia-greek/ionia/parmenides/sullego", homeDir)
			}

			if mode == "" {
				mode = ByFile
			}

			parse(filePath, outDir, mode)

		},
	}
	cmd.PersistentFlags().StringVarP(&filePath, "filepath", "f", "", "where to find the txt file")
	cmd.PersistentFlags().StringVarP(&outDir, "outdir", "o", "", "demokritos dir")
	cmd.PersistentFlags().StringVarP(&mode, "mode", "m", "", "mode for parsing valid options: by-letter, by-file")

	return cmd
}

func parse(filePath, outDir, mode string) {
	plan, _ := os.ReadFile(filePath)
	wordList := strings.Split(string(plan), "\n")
	logging.Info(fmt.Sprintf("found %d words in %s", len(wordList), filePath))

	if mode == ByFile {
		pathParts := strings.Split(filePath, "/")
		name := strings.Split(pathParts[len(pathParts)-1], ".")[0]
		logging.Info(name)

		parseLinesByFile(outDir, name, wordList)
	} else if mode == ByLetter {
		parseLinesByLetter(outDir, wordList)
	} else {
		log.Fatal("No mode provided")
	}
}

func parseLinesByLetter(outDir string, wordList []string) error {
	var biblos models.Biblos
	currentLetter := "α"

	for i, word := range wordList {
		var greek string
		var english string
		for j, char := range word {
			c := fmt.Sprintf("%c", char)
			if j == 0 {
				removedAccent := util.RemoveAccent(c)
				if currentLetter != removedAccent {
					jsonBiblos, err := biblos.Marshal()
					if err != nil {
						return err
					}

					outputFile := fmt.Sprintf("%s/%s.json", outDir, currentLetter)
					util.WriteFile(jsonBiblos, outputFile)
					currentLetter = removedAccent
					biblos = models.Biblos{}
				}
			}
			matched, err := regexp.MatchString(`[A-Za-z]`, c)
			if err != nil {
				return err
			}
			if matched {
				greek = strings.TrimSpace(word[0 : j-1])
				english = strings.TrimSpace(word[j-1:])
				logging.Debug(fmt.Sprintf("found the greek: %s and the english %s", greek, english))

				meros := models.Meros{
					Greek:   greek,
					English: english,
				}

				biblos.Biblos = append(biblos.Biblos, meros)
				break
			}
		}
		if i == len(wordList)-1 {
			jsonBiblos, err := biblos.Marshal()
			if err != nil {
				return err
			}

			outputFile := fmt.Sprintf("%s/%s.json", outDir, currentLetter)
			util.WriteFile(jsonBiblos, outputFile)
		}
	}

	logging.Info(fmt.Sprintf("all words have been parsed and saved to %s", outDir))
	return nil
}

func parseLinesByFile(outDir, name string, wordList []string) error {
	var logos models.Logos

	for _, word := range wordList {
		var greek string
		var translation string
		for j, char := range word {
			c := fmt.Sprintf("%c", char)
			matched, err := regexp.MatchString(`[A-Za-z]`, c)
			if err != nil {
				return err
			}
			if matched {
				greek = strings.TrimSpace(word[0 : j-1])
				translation = strings.TrimSpace(word[j-1:])
				logging.Debug(fmt.Sprintf("found the greek: %s and the translation %s", greek, translation))

				meros := models.Word{
					Greek:       greek,
					Translation: translation,
				}

				logos.Logos = append(logos.Logos, meros)
				break
			}
		}
	}

	numberOfWords := len(logos.Logos)
	var wordsPerChapter int
	switch {
	case numberOfWords < 500:
		wordsPerChapter = 10
	case numberOfWords > 501:
		wordsPerChapter = 20
	}

	chaptersLength := numberOfWords / wordsPerChapter
	lastChapter := chaptersLength + 1
	var randonNumbers []int

	for i := 1; i <= chaptersLength; i++ {
		for j := 1; j <= wordsPerChapter; j++ {
			randomNumber := helpers.GenerateRandomNumber(numberOfWords)
			numberIsUnique := uniqueNumber(randonNumbers, randomNumber)

			for !numberIsUnique {
				randomNumber = helpers.GenerateRandomNumber(numberOfWords)
				numberIsUnique = uniqueNumber(randonNumbers, randomNumber)
			}

			logos.Logos[randomNumber].Chapter = int64(i)
			randonNumbers = append(randonNumbers, randomNumber)
		}
	}

	for i, word := range logos.Logos {
		if word.Chapter == int64(0) {
			logos.Logos[i].Chapter = int64(lastChapter)
		}
	}

	jsonLogos, err := logos.Marshal()
	if err != nil {
		return err
	}

	outputFile := fmt.Sprintf("%s/%s.json", outDir, name)
	util.WriteFile(jsonLogos, outputFile)

	return nil
}

func uniqueNumber(numberList []int, number int) bool {
	numberIsUnique := true
	for _, n := range numberList {
		if n == number {
			numberIsUnique = false

		}
	}

	return numberIsUnique
}
