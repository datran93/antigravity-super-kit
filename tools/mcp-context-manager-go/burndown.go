package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// StepTimestamps is a map of step name → RFC3339 completion timestamp.
type StepTimestamps map[string]string

// StepDrift is a map of step name → failure count.
type StepDrift map[string]int

// ParseStepTimestamps safely unmarshals a JSON string into StepTimestamps.
func ParseStepTimestamps(raw string) StepTimestamps {
	out := make(StepTimestamps)
	if raw == "" || raw == "{}" {
		return out
	}
	json.Unmarshal([]byte(raw), &out) //nolint:errcheck
	return out
}

// ParseStepDrift safely unmarshals a JSON string into StepDrift.
func ParseStepDrift(raw string) StepDrift {
	out := make(StepDrift)
	if raw == "" || raw == "{}" {
		return out
	}
	json.Unmarshal([]byte(raw), &out) //nolint:errcheck
	return out
}

// CalculateVelocity returns the number of steps completed per day.
// Returns 0 if fewer than 2 timestamps are available.
func CalculateVelocity(timestamps StepTimestamps) float64 {
	if len(timestamps) < 2 {
		return 0
	}

	times := make([]time.Time, 0, len(timestamps))
	for _, ts := range timestamps {
		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			continue
		}
		times = append(times, t)
	}

	if len(times) < 2 {
		return 0
	}

	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	span := times[len(times)-1].Sub(times[0])
	days := span.Hours() / 24.0
	if days <= 0 {
		// Steps completed within the same minute: use 1-hour minimum to avoid ÷0
		days = 1.0 / 24.0
	}
	return float64(len(times)) / days
}

// EstimateETA returns a human-readable ETA string given velocity and remaining step count.
func EstimateETA(velocity float64, remainingSteps int) string {
	if remainingSteps == 0 {
		return "done"
	}
	if velocity <= 0 {
		return "unknown"
	}
	days := float64(remainingSteps) / velocity
	if days < 1.0/24.0 {
		minutes := int(math.Round(days * 24 * 60))
		return fmt.Sprintf("~%d min", minutes)
	}
	if days < 1 {
		hours := int(math.Round(days * 24))
		return fmt.Sprintf("~%d hr", hours)
	}
	return fmt.Sprintf("~%.1f days", math.Round(days*10)/10)
}

// driftIcon returns an emoji badge based on the failure count for a step.
//
//	0     → 🟢 (clean)
//	1     → 🟡 (warning)
//	2     → 🟠 (danger)
//	>=3   → 🔴 (blocked)
func driftIcon(count int) string {
	switch {
	case count == 0:
		return "🟢"
	case count == 1:
		return "🟡"
	case count == 2:
		return "🟠"
	default:
		return "🔴"
	}
}

// RenderBurndownHeader generates the sprint header line for progress.md.
// Example: "📊 Sprint: my-task | Velocity: 3.0 steps/day | ETA: ~2 days"
func RenderBurndownHeader(taskID string, timestamps StepTimestamps, remainingSteps int) string {
	vel := CalculateVelocity(timestamps)
	eta := EstimateETA(vel, remainingSteps)
	if vel > 0 {
		return fmt.Sprintf("📊 **Sprint:** `%s` | **Velocity:** %.1f steps/day | **ETA:** %s\n\n", taskID, vel, eta)
	}
	return fmt.Sprintf("📊 **Sprint:** `%s` | **ETA:** %s\n\n", taskID, eta)
}

// RenderStepWithMeta formats a single step line with timestamp and drift badge.
// completed=true  →  "- [x] [P0-T1] Step name  (14:32) 🟢"
// completed=false →  "- [ ] [P1-T1] Step name  🔴 drift:3 BLOCKED"
func RenderStepWithMeta(step string, completed bool, timestamps StepTimestamps, drift StepDrift) string {
	driftCount := drift[step]
	badge := driftIcon(driftCount)

	if completed {
		timeStr := ""
		if ts, ok := timestamps[step]; ok {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				timeStr = fmt.Sprintf("  `(%s)`", t.Format("15:04"))
			}
		}
		suffix := ""
		if driftCount > 0 {
			suffix = fmt.Sprintf(" drift:%d", driftCount)
		}
		return fmt.Sprintf("- [x] %s%s %s%s\n", step, timeStr, badge, suffix)
	}

	// Pending step
	suffix := ""
	if driftCount > 0 {
		suffix = fmt.Sprintf(" drift:%d", driftCount)
		if driftCount >= 3 {
			suffix += " **BLOCKED**"
		}
	}
	return fmt.Sprintf("- [ ] %s %s%s\n", step, badge, suffix)
}

// RenderBurndownSection returns the full burndown-enhanced step section
// to be embedded in progress.md. It replaces the plain checklist rendering.
func RenderBurndownSection(
	completed, pending []string,
	timestamps StepTimestamps,
	drift StepDrift,
) string {
	var sb strings.Builder

	if len(completed) > 0 {
		sb.WriteString("### ✅ Completed\n")
		for _, s := range completed {
			sb.WriteString(RenderStepWithMeta(s, true, timestamps, drift))
		}
		sb.WriteString("\n")
	}

	if len(pending) > 0 {
		sb.WriteString("### ⏳ Next Steps\n")
		for _, s := range pending {
			sb.WriteString(RenderStepWithMeta(s, false, timestamps, drift))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderHistoricallyIncomplete renders a markdown table of stale in_progress tasks
// for embedding in progress.md below the current task's checklist.
func RenderHistoricallyIncomplete(tasks []IdleTask) string {
	if len(tasks) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("---\n### ⚠️ Historically Incomplete Tasks (%d)\n\n", len(tasks)))
	sb.WriteString("| Task ID | Description | Progress | Last Update | Idle |\n")
	sb.WriteString("|---------|-------------|----------|-------------|------|\n")
	for _, t := range tasks {
		desc := t.Description
		if len(desc) > 50 {
			desc = desc[:47] + "…"
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %.0f%% (%d/%d) | %s | %d days |\n",
			t.TaskID, desc, t.Progress, t.Done, t.Total, t.LastUpdate, t.IdleDays))
	}
	sb.WriteString("\n> Run `load_checkpoint(task_id=\"<id>\")` to resume any of these tasks.\n\n")
	return sb.String()
}
