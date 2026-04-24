package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TransformWorkflows converts all workflows from source to agent-specific format.
func TransformWorkflows(agent *AgentEntry, sourceDir, targetBaseDir string) error {
	wf := agent.Workflows
	if wf == nil {
		return nil
	}

	outputDir := filepath.Join(targetBaseDir, wf.Dir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	srcDir := filepath.Join(sourceDir, "workflows")
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		srcFile := filepath.Join(srcDir, entry.Name())
		baseName := strings.TrimSuffix(entry.Name(), ".md")

		switch wf.Format {
		case "toml":
			outFile := filepath.Join(outputDir, baseName+wf.Extension)
			if err := transformToTOML(srcFile, outFile, agent.ArgsPlaceholder); err != nil {
				return fmt.Errorf("TOML transform %s: %w", entry.Name(), err)
			}
		case "skill_folder":
			if err := transformToSkillFolder(srcFile, outputDir); err != nil {
				return fmt.Errorf("skill folder %s: %w", entry.Name(), err)
			}
		case "md":
			if wf.Extension == ".agent.md" {
				outFile := filepath.Join(outputDir, baseName+".agent.md")
				if err := copyFile(srcFile, outFile); err != nil {
					return err
				}
			} else {
				outFile := filepath.Join(outputDir, baseName+wf.Extension)
				if err := transformMD(srcFile, outFile, agent.ArgsPlaceholder); err != nil {
					return err
				}
			}
		default:
			// Unknown format — copy verbatim
			outFile := filepath.Join(outputDir, entry.Name())
			if err := copyFile(srcFile, outFile); err != nil {
				return err
			}
		}
	}
	return nil
}

// TransformRules copies and adapts rules for the target agent.
func TransformRules(agent *AgentEntry, sourceDir, targetBaseDir string) error {
	rules := agent.Rules
	if rules == nil {
		return nil
	}

	outputDir := filepath.Join(targetBaseDir, rules.Dir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	srcDir := filepath.Join(sourceDir, "rules")
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcFile := filepath.Join(srcDir, entry.Name())
		targetName := entry.Name()

		// Apply rename map
		if rules.Rename != nil {
			if newName, ok := rules.Rename[entry.Name()]; ok {
				targetName = newName
			}
		}

		// Change extension if needed
		if rules.Extension != "" && rules.Extension != ".md" {
			targetName = strings.TrimSuffix(targetName, filepath.Ext(targetName)) + rules.Extension
		}

		outFile := filepath.Join(outputDir, targetName)
		if err := copyFile(srcFile, outFile); err != nil {
			return err
		}
	}
	return nil
}

// CopySkills copies skill folders to target if agent supports skills.
func CopySkills(agent *AgentEntry, sourceDir, targetBaseDir string) error {
	if agent.Skills == nil {
		return nil
	}

	srcDir := filepath.Join(sourceDir, "skills")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil
	}

	outputDir := filepath.Join(targetBaseDir, agent.Skills.Dir)
	return copyDirRecursive(srcDir, outputDir)
}

// CopyReferences copies reference files to target if agent supports them.
func CopyReferences(agent *AgentEntry, sourceDir, targetBaseDir string) error {
	if agent.References == nil {
		return nil
	}

	srcDir := filepath.Join(sourceDir, "references")
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil
	}

	outputDir := filepath.Join(targetBaseDir, agent.References.Dir)
	return copyDirRecursive(srcDir, outputDir)
}

// InstallAgent runs the full installation pipeline for a single agent.
func InstallAgent(agent *AgentEntry, sourceAgentsDir, projectDir string) error {
	targetDir := filepath.Join(projectDir, agent.TargetDir)

	if agent.Verbatim {
		return copyDirRecursive(sourceAgentsDir, targetDir)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	if err := TransformRules(agent, sourceAgentsDir, targetDir); err != nil {
		return fmt.Errorf("rules: %w", err)
	}
	if err := TransformWorkflows(agent, sourceAgentsDir, targetDir); err != nil {
		return fmt.Errorf("workflows: %w", err)
	}
	if err := CopySkills(agent, sourceAgentsDir, targetDir); err != nil {
		return fmt.Errorf("skills: %w", err)
	}
	if err := CopyReferences(agent, sourceAgentsDir, targetDir); err != nil {
		return fmt.Errorf("references: %w", err)
	}
	return nil
}

// ── Transform helpers ───────────────────────────────────────────────────────

// transformToTOML converts YAML-frontmatter Markdown to Gemini TOML format.
func transformToTOML(inputFile, outputFile, argsPlaceholder string) error {
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	text := string(content)
	description := ""
	body := text

	// Parse YAML frontmatter
	if strings.HasPrefix(text, "---") {
		parts := strings.SplitN(text, "---", 3)
		if len(parts) >= 3 {
			frontmatter := strings.TrimSpace(parts[1])
			body = strings.TrimSpace(parts[2])
			for _, line := range strings.Split(frontmatter, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "description:") {
					desc := strings.TrimPrefix(line, "description:")
					desc = strings.TrimSpace(desc)
					desc = strings.Trim(desc, `"'`)
					description = desc
				}
			}
		}
	}

	// Replace $ARGUMENTS with agent placeholder
	body = strings.ReplaceAll(body, "$ARGUMENTS", argsPlaceholder)

	toml := fmt.Sprintf("description = %q\n\n%s\n", description, body)
	return os.WriteFile(outputFile, []byte(toml), 0644)
}

// transformMD converts a Markdown workflow for MD-based agents.
func transformMD(inputFile, outputFile, argsPlaceholder string) error {
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	text := string(content)
	if argsPlaceholder != "$ARGUMENTS" {
		text = strings.ReplaceAll(text, "$ARGUMENTS", argsPlaceholder)
	}

	return os.WriteFile(outputFile, []byte(text), 0644)
}

// transformToSkillFolder creates a skill folder with SKILL.md.
func transformToSkillFolder(inputFile, outputDir string) error {
	baseName := strings.TrimSuffix(filepath.Base(inputFile), ".md")
	skillDir := filepath.Join(outputDir, baseName)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}
	return copyFile(inputFile, filepath.Join(skillDir, "SKILL.md"))
}

// ── File helpers ────────────────────────────────────────────────────────────

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func copyDirRecursive(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != src {
			return filepath.SkipDir
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

// CountFiles counts regular files under a directory.
func CountFiles(dir string) int {
	count := 0
	filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}
