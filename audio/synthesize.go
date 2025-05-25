package audio

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/lafeingcrokodil/flashcards/io"
)

// Config configures how to generate MP3 files from text in a CSV file.
type Config struct {
	// Client is a text to speech client.
	Client *texttospeech.Client
	// CSVPath is the path to the CSV file containing the text to be converted to speech.
	CSVPath string
	// Delimiter is the character that separates values in the file.
	Delimiter rune
	// ColumnName is the name of the column containing the text to be converted to speech.
	ColumnName string
	// LanguageCode is the xx-XX code for the language that the text is in.
	LanguageCode string
	// OutputDir is the directory to which the MP3 files should be written.
	OutputDir string
}

// CreateMP3FromCSV creates MP3 files by converting text in a CSV file to speech.
func CreateMP3FromCSV(ctx context.Context, cfg Config) error {
	records, err := io.ReadCSVFile(cfg.CSVPath, cfg.Delimiter)
	if err != nil {
		return err
	}

	for _, record := range records {
		err = CreateMP3(ctx, record[cfg.ColumnName], cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateMP3FromCSV creates an MP3 file by converting the specified text to speech.
// If the MP3 file already exists, this function does nothing.
func CreateMP3(ctx context.Context, text string, cfg Config) error {
	outputPath := path.Join(cfg.OutputDir, fmt.Sprintf("%s.mp3", text))

	_, err := os.Stat(outputPath)
	if !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("INFO\t%s already exists; skipping...\n", outputPath)
		return nil
	}

	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: cfg.LanguageCode,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := cfg.Client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, resp.AudioContent, 0644)
}
