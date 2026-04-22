package main

import (
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
