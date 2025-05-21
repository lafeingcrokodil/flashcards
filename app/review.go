package app

import (
	"bufio"
	"fmt"
	"os"
)

// Review presents the specified flashcards to the user.
func Review(fcs []*Flashcard) error {
	r := bufio.NewReader(os.Stdin)

nextCard:
	for _, fc := range fcs {
		prompt := fc.Prompt
		if fc.Context != "" {
			prompt += fmt.Sprintf(" (%s)", fc.Context)
		}
		fmt.Printf("%s\n> ", prompt)
		answer, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		for _, expected := range fc.Answers {
			if answer == fmt.Sprintf("%s\n", expected) {
				fmt.Print("Correct!\n\n")
				continue nextCard
			}
		}
		fmt.Printf("The expected answer was: %s\n\n", fc.Answers[0])
	}

	return nil
}
