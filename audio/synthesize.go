package audio

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

// InputReader reads input text for which audio should be generated.
type InputReader interface {
	// Read returns a list of strings for which audio should be generated.
	Read(ctx context.Context) ([]string, error)
}

// Synthesizer generates MP3 files from text.
type Synthesizer struct {
	// Client is a text to speech client.
	Client *texttospeech.Client
	// LanguageCode is the xx-XX code for the language that the text is in.
	LanguageCode string
	// OutputDir is the directory to which the MP3 files should be written.
	OutputDir string
}

// CreateMP3s creates MP3 files by converting input text to speech.
func (s *Synthesizer) CreateMP3s(ctx context.Context, r InputReader) error {
	inputs, err := r.Read(ctx)
	if err != nil {
		return err
	}

	for _, input := range inputs {
		err = s.createMP3(ctx, input)
		if err != nil {
			return err
		}
	}

	return nil
}

// createMP3 creates an MP3 file by converting the specified text to speech.
// If the MP3 file already exists, this function does nothing.
func (s *Synthesizer) createMP3(ctx context.Context, text string) error {
	outputPath := path.Join(s.OutputDir, fmt.Sprintf("%s.mp3", text))

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
			LanguageCode: s.LanguageCode,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_MALE,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := s.Client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, resp.AudioContent, 0644)
}
