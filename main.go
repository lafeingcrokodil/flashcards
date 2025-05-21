package main

import (
	"bufio"
	"fmt"
	"os"
)

// Flashcard stores the expected answer for a given prompt.
type Flashcard struct {
	// Prompt is the text to be shown to the user.
	Prompt string
	// Context helps narrow down possible answers.
	Context string
	// Answer is the set of accepted answers.
	Answers []string
}

var cards = []*Flashcard{
	{
		Prompt:  "Hello",
		Context: "informal, polite",
		Answers: []string{"안녕하세요"},
	},
	{
		Prompt:  "Sunday",
		Answers: []string{"일요일"},
	},
}

func main() {
	r := bufio.NewReader(os.Stdin)

nextCard:
	for _, c := range cards {
		prompt := c.Prompt
		if c.Context != "" {
			prompt += fmt.Sprintf(" (%s)", c.Context)
		}
		fmt.Printf("%s\n> ", prompt)
		answer, err := r.ReadString('\n')
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		for _, expected := range c.Answers {
			if answer == fmt.Sprintf("%s\n", expected) {
				fmt.Print("Correct!\n\n")
				continue nextCard
			}
		}
		fmt.Printf("The expected answer was: %s\n\n", c.Answers[0])
	}
}
