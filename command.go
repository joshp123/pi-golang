package pi

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Command struct {
	Executable string
	Args       []string
}

func (command Command) WithArgs(extra []string) []string {
	args := append([]string{}, command.Args...)
	args = append(args, extra...)
	return args
}

func ResolveCommand() (Command, error) {
	piPath, err := exec.LookPath("pi")
	if err != nil {
		return Command{}, errors.New("pi CLI not found; install @mariozechner/pi-coding-agent")
	}
	if cmd, ok := commandFromPiWrapper(piPath); ok {
		return cmd, nil
	}
	return Command{Executable: piPath}, nil
}

func commandFromPiWrapper(piPath string) (Command, bool) {
	if resolved := resolveCliFromWrapper(piPath); resolved.Executable != "" {
		return resolved, true
	}

	root := filepath.Dir(filepath.Dir(piPath))
	candidate := filepath.Join(root, "lib", "node_modules", "@mariozechner", "pi-coding-agent", "dist", "cli.js")
	if fileExists(candidate) {
		return Command{Executable: "node", Args: []string{candidate}}, true
	}

	fallback := filepath.Join(root, "lib", "node_modules", "pi-monorepo", "dist", "cli.js")
	if fileExists(fallback) {
		return Command{Executable: "node", Args: []string{fallback}}, true
	}

	return Command{}, false
}

func resolveCliFromWrapper(piPath string) Command {
	file, err := os.Open(piPath)
	if err != nil {
		return Command{}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "exec ") {
			continue
		}
		fields := strings.Fields(line)
		var nodePath string
		var cliPath string
		for _, field := range fields {
			clean := strings.Trim(field, "\"")
			if strings.HasSuffix(clean, "node") && fileExists(clean) {
				nodePath = clean
			}
			if strings.HasSuffix(clean, "cli.js") && fileExists(clean) {
				cliPath = clean
			}
		}
		if cliPath != "" {
			if nodePath == "" {
				nodePath = "node"
			}
			return Command{Executable: nodePath, Args: []string{cliPath}}
		}
	}
	return Command{}
}
