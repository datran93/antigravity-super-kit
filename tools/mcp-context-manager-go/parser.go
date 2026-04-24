package main

import (
	"os"
	"regexp"
	"strings"
)

type StepInfo struct {
	ID           string
	Name         string
	Dependencies []string
}

// ParseStep parses a step string into StepInfo.
// Expected format: "[ST01][core] Description depends:[ST02,ST03]"
func ParseStep(stepStr string) StepInfo {
	info := StepInfo{
		Name: stepStr, // Keep full string as the name for context
	}

	// Extract step ID, e.g., [ST01]
	idRe := regexp.MustCompile(`^\[([A-Za-z0-9_-]+)\]`)
	idMatch := idRe.FindStringSubmatch(stepStr)
	if len(idMatch) > 1 {
		info.ID = idMatch[1]
	}

	// Extract dependencies, e.g., depends:[ST02,ST03]
	depRe := regexp.MustCompile(`depends:\[(.*?)\]`)
	depMatch := depRe.FindStringSubmatch(stepStr)
	if len(depMatch) > 1 {
		depsStr := depMatch[1]
		parts := strings.Split(depsStr, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				info.Dependencies = append(info.Dependencies, p)
			}
		}
	}

	return info
}

// ExtractLinks extracts @ki:[id] and @task:[id] from a text.
func ExtractLinks(text string) (kis []string, tasks []string) {
	kiRe := regexp.MustCompile(`@ki:\[([^\]]+)\]`)
	taskRe := regexp.MustCompile(`@task:\[([^\]]+)\]`)

	for _, match := range kiRe.FindAllStringSubmatch(text, -1) {
		if len(match) > 1 {
			kis = append(kis, strings.TrimSpace(match[1]))
		}
	}

	for _, match := range taskRe.FindAllStringSubmatch(text, -1) {
		if len(match) > 1 {
			tasks = append(tasks, strings.TrimSpace(match[1]))
		}
	}

	return kis, tasks
}

// ExtractAcceptanceCriteria parses a spec.md file and extracts the Acceptance Criteria section.
func ExtractAcceptanceCriteria(specPath string) string {
	content, err := os.ReadFile(specPath)
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`(?i)##\s+Acceptance Criteria\s*\n([\s\S]*?)(?:\n##\s+|$)`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
