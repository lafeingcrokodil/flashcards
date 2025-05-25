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
	defer c.Close()

	cfg := audio.Config{
		Client:       c,
		CSVPath:      "data/translations.tsv",
		Delimiter:    '\t',
		ColumnName:   "korean",
		LanguageCode: "ko-KR",
		OutputDir:    "public/audio",
	}

	return audio.CreateMP3FromCSV(ctx, cfg)
}
