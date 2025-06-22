package main

import (
	"context"
	"fmt"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/lafeingcrokodil/flashcards/audio"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("ERROR\t%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	c, err := texttospeech.NewClient(ctx)
	if err != nil {
		return err
	}
	defer c.Close() // nolint:errcheck

	r := &audio.CSVReader{
		CSVPath:    "data/translations.tsv",
		Delimiter:  '\t',
		ColumnName: "korean",
	}

	s := &audio.Synthesizer{
		Client:       c,
		LanguageCode: "ko-KR",
		OutputDir:    "public/audio",
	}

	return s.CreateMP3s(ctx, r)
}
