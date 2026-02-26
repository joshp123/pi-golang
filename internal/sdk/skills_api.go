package sdk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (client *Client) ListLoadedSkills(ctx context.Context) ([]LoadedSkill, error) {
	commands, err := client.getCommands(ctx)
	if err != nil {
		return nil, err
	}

	skills := make([]LoadedSkill, 0, len(commands))
	for _, command := range commands {
		if strings.TrimSpace(command.Source) != "skill" {
			continue
		}
		name := strings.TrimSpace(command.Name)
		name = strings.TrimPrefix(name, "skill:")
		skills = append(skills, LoadedSkill{
			Name:        name,
			Description: strings.TrimSpace(command.Description),
			Path:        strings.TrimSpace(command.Path),
			Location:    toSkillLocation(command.Location),
		})
	}

	return skills, nil
}

func (client *Client) getCommands(ctx context.Context) ([]slashCommand, error) {
	response, err := client.send(ctx, getCommandsCommand())
	if err != nil {
		return nil, err
	}
	return decodeCommands(response.Data)
}

func (client *Client) verifyLoadedSkills(ctx context.Context, explicitPaths []string) error {
	allowed, err := buildAllowedSkillPaths(explicitPaths)
	if err != nil {
		return err
	}

	skills, err := client.ListLoadedSkills(ctx)
	if err != nil {
		return fmt.Errorf("list loaded skills: %w", err)
	}
	if len(skills) == 0 {
		return fmt.Errorf("explicit skills mode loaded no skills")
	}

	matched := make([]bool, len(allowed))
	for _, skill := range skills {
		skillPath := strings.TrimSpace(skill.Path)
		if skillPath == "" {
			return fmt.Errorf("loaded skill %q missing source path", skill.Name)
		}
		resolvedSkillPath, err := canonicalPath(skillPath)
		if err != nil {
			return fmt.Errorf("resolve loaded skill path %q: %w", skillPath, err)
		}

		pathMatched := false
		for index, entry := range allowed {
			if skillPathMatches(resolvedSkillPath, entry) {
				matched[index] = true
				pathMatched = true
				break
			}
		}
		if !pathMatched {
			return fmt.Errorf("unexpected loaded skill %q from %q (outside explicit skills paths)", skill.Name, skillPath)
		}
	}

	for index, wasMatched := range matched {
		if !wasMatched {
			return fmt.Errorf("explicit skill path %q loaded no skills", allowed[index].path)
		}
	}

	return nil
}

type allowedSkillPath struct {
	path  string
	isDir bool
}

func buildAllowedSkillPaths(paths []string) ([]allowedSkillPath, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("explicit skills mode requires at least one skill path")
	}

	allowed := make([]allowedSkillPath, 0, len(paths))
	seen := map[string]struct{}{}
	for _, raw := range paths {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		resolvedPath, err := canonicalPath(path)
		if err != nil {
			return nil, fmt.Errorf("resolve explicit skill path %q: %w", path, err)
		}
		if _, exists := seen[resolvedPath]; exists {
			continue
		}
		info, err := os.Stat(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("stat explicit skill path %q: %w", resolvedPath, err)
		}
		allowed = append(allowed, allowedSkillPath{path: resolvedPath, isDir: info.IsDir()})
		seen[resolvedPath] = struct{}{}
	}
	if len(allowed) == 0 {
		return nil, fmt.Errorf("explicit skills mode requires at least one skill path")
	}
	return allowed, nil
}

func canonicalPath(path string) (string, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if symlinkResolved, err := filepath.EvalSymlinks(resolved); err == nil {
		resolved = symlinkResolved
	}
	return filepath.Clean(resolved), nil
}

func skillPathMatches(skillPath string, entry allowedSkillPath) bool {
	if !entry.isDir {
		return skillPath == entry.path
	}
	if skillPath == entry.path {
		return true
	}
	prefix := entry.path
	if !strings.HasSuffix(prefix, string(os.PathSeparator)) {
		prefix += string(os.PathSeparator)
	}
	return strings.HasPrefix(skillPath, prefix)
}

func toSkillLocation(value string) SkillLocation {
	switch strings.TrimSpace(value) {
	case string(SkillLocationUser):
		return SkillLocationUser
	case string(SkillLocationProject):
		return SkillLocationProject
	case string(SkillLocationPath):
		return SkillLocationPath
	default:
		return SkillLocationUnknown
	}
}
