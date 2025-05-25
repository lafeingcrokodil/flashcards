package review

import (
	"fmt"
	"slices"

	"github.com/lafeingcrokodil/flashcards/io"
)

// Flashcard stores the expected answer for a given prompt.
type Flashcard struct {
	// Prompt is the text to be shown to the user.
	Prompt string `json:"prompt"`
	// Context helps narrow down possible answers.
	Context string `json:"context"`
	// Answer is the accepted answer.
	Answer string `json:"answer"`
	// Proficiency indicates how reliably the user provides the correct answer.
	Proficiency int `json:"proficiency"`
	// ViewCount is the number of times the flashcard has been reviewed.
	ViewCount int `json:"viewCount"`
}

// LoadConfig configures how to load flashcards from a file.
type LoadConfig struct {
	// Filepath is the path to the file.
	Filepath string
	// Delimiter is the character that separates values in the file.
	Delimiter rune
	// PromptHeader is the name of the column containing the prompts.
	PromptHeader string
	// ContextHeader is the name of the column containing the context (if any).
	ContextHeader string
	// AnswerHeader is the name of the column containing the answers.
	AnswerHeader string
}

// qualifiedPrompt returns a prompt together with the context (if there is one).
func qualifiedPrompt(prompt, context string) string {
	if context != "" {
		return fmt.Sprintf("%s (%s)", prompt, context)
	}
	return prompt
}

// Check returns true if the provided answer matches the expected one.
func (f *Flashcard) Check(answer string) bool {
	return answer == f.Answer
}

// LoadFromCSV loads flashcards from a CSV file.
func LoadFromCSV(lc LoadConfig) ([]*Flashcard, error) {
	var fs []*Flashcard

	// Load records from a CSV file.
	records, err := io.ReadCSVFile(lc.Filepath, lc.Delimiter)
	if err != nil {
		return nil, err
	}

	// Group CSV records by the qualified prompt (prompt + context).
	recordsByPrompt := make(map[string][]map[string]string, len(records))
	for _, record := range records {
		prompt := qualifiedPrompt(record[lc.PromptHeader], record[lc.ContextHeader])
		if rs, ok := recordsByPrompt[prompt]; ok {
			recordsByPrompt[prompt] = append(rs, record)
		} else {
			recordsByPrompt[prompt] = []map[string]string{record}
		}
	}

	// Get the alphabetically sorted list of qualified prompts.
	var prompts []string
	for prompt := range recordsByPrompt {
		prompts = append(prompts, prompt)
	}
	slices.Sort(prompts)

	// Create a flashcard for each qualified prompt.
	for _, prompt := range prompts {
		var f *Flashcard
		records := recordsByPrompt[prompt]
		for _, record := range records {
			if f == nil {
				f = &Flashcard{
					Prompt:  record[lc.PromptHeader],
					Context: record[lc.ContextHeader],
					Answer:  record[lc.AnswerHeader],
				}
			} else {
				return nil, fmt.Errorf("ambiguous answer for %s", qualifiedPrompt(f.Prompt, f.Context))
			}
		}
		fs = append(fs, f)
	}

	return fs, nil
}
