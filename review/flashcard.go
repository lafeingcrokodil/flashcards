package review

import (
	"fmt"
)

// Flashcard stores the expected answer for a given prompt.
type Flashcard struct {
	// ID uniquely identifies a flashcard.
	ID string `json:"id"`
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

// Check returns true if the provided answer matches the expected one.
func (f *Flashcard) Check(answer string) bool {
	return answer == f.Answer
}

// QualifiedPrompt returns the flashcard prompt together with the context (if there is one).
func (f *Flashcard) QualifiedPrompt() string {
	if f.Context != "" {
		return fmt.Sprintf("%s (%s)", f.Prompt, f.Context)
	}
	return f.Prompt
}
