package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// depsSuffix is the keyword used in step names to declare dependencies.
// Example step: "[P1-T1] Build pipeline depends:[P0-T1,P0-T2]"
const depsSuffix = " depends:["

// ParseStepDeps extracts the dependency map from a list of steps.
// Steps with a "depends:[P0-T1,P0-T2]" suffix are parsed.
// Returns map[stepBaseName][]depLabel and the cleaned step name (suffix stripped).
func ParseStepDeps(steps []string) (cleanedSteps []string, deps map[string][]string) {
	deps = make(map[string][]string)
	cleanedSteps = make([]string, 0, len(steps))

	for _, s := range steps {
		idx := strings.Index(s, depsSuffix)
		if idx < 0 {
			cleanedSteps = append(cleanedSteps, s)
			continue
		}
		base := strings.TrimSpace(s[:idx])
		depsSection := s[idx+len(depsSuffix):]
		// Strip closing ']'
		closeBracket := strings.LastIndex(depsSection, "]")
		if closeBracket >= 0 {
			depsSection = depsSection[:closeBracket]
		}
		var depList []string
		for _, d := range strings.Split(depsSection, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				depList = append(depList, d)
			}
		}
		deps[base] = depList
		cleanedSteps = append(cleanedSteps, base)
	}
	return cleanedSteps, deps
}

// IsParallelReady returns true if all dependencies of the given step are in completedSet.
func IsParallelReady(step string, deps map[string][]string, completedSet map[string]bool) bool {
	depList, hasDeps := deps[step]
	if !hasDeps || len(depList) == 0 {
		return true // No deps → always ready
	}
	for _, d := range depList {
		if !completedSet[d] {
			return false
		}
	}
	return true
}

// HasCycle detects circular dependencies using DFS.
// Returns true if a cycle exists.
func HasCycle(deps map[string][]string) bool {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		inStack[node] = true
		for _, dep := range deps[node] {
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			} else if inStack[dep] {
				return true
			}
		}
		inStack[node] = false
		return false
	}

	for node := range deps {
		if !visited[node] {
			if dfs(node) {
				return true
			}
		}
	}
	return false
}

// labelToMermaidID converts a step label like "[P0-T1]" to a Mermaid node ID like "P0T1".
func labelToMermaidID(label string) string {
	id := strings.TrimPrefix(label, "[")
	id = strings.TrimSuffix(id, "]")
	// Strip the tactic part: "[P0-T1] Build X" → "P0T1"
	if spaceIdx := strings.Index(id, " "); spaceIdx >= 0 {
		id = id[:spaceIdx]
	}
	id = strings.ReplaceAll(id, "-", "")
	return id
}

// stepShortLabel extracts the phase prefix for the Mermaid node display.
// "[P0-T1] Build migration" → "P0-T1"
func stepShortLabel(step string) string {
	if !strings.HasPrefix(step, "[") {
		return step
	}
	end := strings.Index(step, "]")
	if end < 0 {
		return step
	}
	return step[1:end]
}

// completedLabel adds an icon suffix to the node label.
func completedLabel(step string, completedSet map[string]bool) string {
	label := stepShortLabel(step)
	if completedSet[step] {
		return label + " ✅"
	}
	return label + " ⏳"
}

// RenderDAGBlock renders the dependency graph as a Mermaid LR diagram.
// Only emits the diagram if there are actual dependencies to show.
func RenderDAGBlock(allSteps []string, deps map[string][]string, completedSet map[string]bool) string {
	if len(deps) == 0 {
		return ""
	}

	if HasCycle(deps) {
		return "> ⚠️ **Dependency cycle detected** — fix `depends:[...]` declarations before rendering DAG.\n\n"
	}

	var sb strings.Builder
	sb.WriteString("### 🔗 Dependency Graph\n\n")
	sb.WriteString("```mermaid\ngraph LR\n")

	// Collect known nodes (steps that appear in deps as key or value)
	nodeSet := make(map[string]bool)
	for step := range deps {
		nodeSet[step] = true
		for _, d := range deps[step] {
			nodeSet[d] = true
		}
	}

	// Emit node definitions with status labels
	for _, step := range allSteps {
		if nodeSet[step] {
			id := labelToMermaidID(step)
			label := completedLabel(step, completedSet)
			sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", id, label))
		}
	}

	// Emit edges
	for step, depList := range deps {
		toID := labelToMermaidID(step)
		for _, dep := range depList {
			fromID := labelToMermaidID(dep)
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", fromID, toID))
		}
	}

	sb.WriteString("```\n\n")
	return sb.String()
}

// BuildCompletedSet creates a set from the completed steps slice for O(1) lookup.
func BuildCompletedSet(completed []string) map[string]bool {
	s := make(map[string]bool, len(completed))
	for _, c := range completed {
		s[c] = true
	}
	return s
}

// GetTaskDAG loads a checkpoint and renders its step dependency graph as Mermaid.
// Returns a human-readable block showing completed (✅) and pending (⏳) steps.
func GetTaskDAG(workspacePath, taskID string) (string, error) {
	db, err := GetDBConnection(workspacePath)
	if err != nil {
		return "", err
	}
	defer db.Close()

	row := db.QueryRow(
		"SELECT completed_steps, next_steps FROM checkpoints WHERE task_id = ?",
		taskID,
	)
	var completedStr, nextStr string
	if err := row.Scan(&completedStr, &nextStr); err != nil {
		return fmt.Sprintf("❌ Task '%s' not found.", taskID), nil
	}

	var comp, nxt []string
	json.Unmarshal([]byte(completedStr), &comp)
	json.Unmarshal([]byte(nextStr), &nxt)

	allSteps := append(comp, nxt...)
	cleanedSteps, deps := ParseStepDeps(allSteps)
	completedSet := BuildCompletedSet(comp)

	if len(deps) == 0 {
		// No dependencies declared — show flat step list instead
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("🔗 Task DAG: %s (no dependencies declared)\n\n", taskID))
		for _, s := range cleanedSteps {
			icon := "⏳"
			if completedSet[s] {
				icon = "✅"
			}
			sb.WriteString(fmt.Sprintf("  %s %s\n", icon, s))
		}
		return sb.String(), nil
	}

	dag := RenderDAGBlock(cleanedSteps, deps, completedSet)
	return fmt.Sprintf("🔗 Task DAG: %s\n\n%s", taskID, dag), nil
}
