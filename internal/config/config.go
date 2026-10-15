package config

import (
	"os"
	"path/filepath"
)

// allowed providers
const (
	GroqBaseURL       = "https://api.groq.com/openai/v1"
	OpenrouterBaseURL = "https://openrouter.ai/api/v1"
)

var nuvoraPath string

func Init() error {
	home, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	nuvoraPath = filepath.Join(home, ".nuvora")

	err = os.MkdirAll(nuvoraPath, 0755)

	if err != nil {
		return err
	}

	return nil
}

func GetNuvoraPath() string {
	return nuvoraPath
}
