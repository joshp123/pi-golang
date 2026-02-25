package pi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Batteries layer: convenience session-sharing helper on top of ExportHTML.

func (client *SessionClient) ShareSession(ctx context.Context) (ShareResult, error) {
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return ShareResult{}, errors.New("gh CLI not found; install https://cli.github.com/")
	}

	tmpDir, err := os.MkdirTemp("", "pi-golang-session-")
	if err != nil {
		return ShareResult{}, err
	}
	defer os.RemoveAll(tmpDir)

	exportPath := filepath.Join(tmpDir, "session.html")
	path, err := client.ExportHTML(ctx, exportPath)
	if err != nil {
		return ShareResult{}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, ghPath, "gist", "create", "--public=false", path)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText == "" {
			errText = err.Error()
		}
		return ShareResult{}, fmt.Errorf("gh gist create failed: %s", errText)
	}

	gistURL := strings.TrimSpace(stdout.String())
	if gistURL == "" {
		return ShareResult{}, errors.New("gh gist create returned empty URL")
	}

	parts := strings.Split(strings.TrimRight(gistURL, "/"), "/")
	gistID := parts[len(parts)-1]
	if strings.TrimSpace(gistID) == "" {
		return ShareResult{}, errors.New("failed to parse gist ID")
	}

	previewURL := fmt.Sprintf("https://shittycodingagent.ai/session/?%s", gistID)
	return ShareResult{GistURL: gistURL, GistID: gistID, PreviewURL: previewURL}, nil
}
