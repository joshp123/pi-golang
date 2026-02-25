package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	pi "github.com/joshp123/pi-golang"
)

func main() {
	auth, mode, err := authFromEnvironment()
	if err != nil {
		log.Fatalf("auth: %v", err)
	}
	if mode == "" {
		log.Println("set ANTHROPIC_API_KEY[/_FILE] or OPENAI_API_KEY[/_FILE] to run this example")
		return
	}

	opts := pi.DefaultOneShotOptions()
	opts.Mode = mode
	opts.Auth = auth
	opts.SeedAuthFromHome = false
	if path := strings.TrimSpace(os.Getenv("PATH")); path != "" {
		opts.Environment = map[string]string{"PATH": path}
	}

	client, err := pi.StartOneShot(opts)
	if err != nil {
		log.Fatalf("start: %v", err)
	}
	defer client.Close()

	result, err := client.Run(context.Background(), pi.PromptRequest{Message: "Say hello in one sentence."})
	if err != nil {
		log.Fatalf("run: %v", err)
	}

	fmt.Println(result.Text)
}

func authFromEnvironment() (pi.ProviderAuth, pi.Mode, error) {
	auth := pi.ProviderAuth{}

	anthropicKey, anthropicKeySet, err := credentialFromEnv("ANTHROPIC_API_KEY")
	if err != nil {
		return pi.ProviderAuth{}, "", err
	}
	anthropicOAuth, anthropicOAuthSet, err := credentialFromEnv("ANTHROPIC_OAUTH_TOKEN")
	if err != nil {
		return pi.ProviderAuth{}, "", err
	}
	anthropicTokenFile := strings.TrimSpace(os.Getenv("ANTHROPIC_TOKEN_FILE"))

	if anthropicKeySet {
		auth.Anthropic.APIKey = anthropicKey
	}
	if anthropicOAuthSet {
		auth.Anthropic.OAuthToken = anthropicOAuth
	}
	if anthropicTokenFile != "" {
		auth.Anthropic.TokenFilePath = anthropicTokenFile
	}
	if anthropicKeySet || anthropicOAuthSet || anthropicTokenFile != "" {
		return auth, pi.ModeSmart, nil
	}

	openAIKey, openAIKeySet, err := credentialFromEnv("OPENAI_API_KEY")
	if err != nil {
		return pi.ProviderAuth{}, "", err
	}
	if openAIKeySet {
		auth.OpenAI.APIKey = openAIKey
		return auth, pi.ModeCoding, nil
	}

	return pi.ProviderAuth{}, "", nil
}

func credentialFromEnv(name string) (pi.Credential, bool, error) {
	value := strings.TrimSpace(os.Getenv(name))
	filePath := strings.TrimSpace(os.Getenv(name + "_FILE"))

	if value != "" && filePath != "" {
		return pi.Credential{}, false, fmt.Errorf("%s and %s_FILE both set", name, name)
	}
	if value != "" {
		return pi.Credential{Value: value}, true, nil
	}
	if filePath == "" {
		return pi.Credential{}, false, nil
	}
	if err := validateCredentialFile(filePath); err != nil {
		return pi.Credential{}, false, fmt.Errorf("%s_FILE: %w", name, err)
	}
	return pi.Credential{File: filePath}, true, nil
}

func validateCredentialFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("path points to directory")
	}
	if info.Size() == 0 {
		return errors.New("file is empty")
	}
	return nil
}
