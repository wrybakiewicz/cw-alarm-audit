package main

import (
	"fmt"
	"sort"
	"time"
)

// isProblematic checks if an alarm is problematic
func isProblematic(r row) bool {
	// Alarm is problematic if:
	// 1. State is ALARM or INSUFFICIENT_DATA
	// 2. Actions are disabled
	// 3. No actions configured at all
	if r.State == "ALARM" || r.State == "INSUFFICIENT_DATA" {
		return true
	}
	if !r.ActionsEnabled {
		return true
	}
	if r.AlarmActions == 0 && r.OKActions == 0 && r.InsufActions == 0 {
		return true
	}
	return false
}

// filterOptions contains all filter options
type filterOptions struct {
	onlyBroken      bool
	stateFilter     string
	noActions       bool
	actionsDisabled bool
	stale           time.Duration
}

// applyFilters applies all filters to the rows and returns filtered results
func applyFilters(rows []row, opts filterOptions) []row {
	now := time.Now()
	var filtered []row

	for _, r := range rows {
		// Filter by broken/problematic
		if opts.onlyBroken && !isProblematic(r) {
			continue
		}

		// Filter by state
		if opts.stateFilter != "" && r.State != opts.stateFilter {
			continue
		}

		// Filter by no actions
		if opts.noActions && (r.AlarmActions != 0 || r.OKActions != 0 || r.InsufActions != 0) {
			continue
		}

		// Filter by actions disabled
		if opts.actionsDisabled && r.ActionsEnabled {
			continue
		}

		// Filter by stale
		if opts.stale > 0 {
			if r.StateUpdatedTimestamp == nil {
				continue
			}
			timeSinceUpdate := now.Sub(*r.StateUpdatedTimestamp)
			if timeSinceUpdate < opts.stale {
				continue
			}
		}

		filtered = append(filtered, r)
	}

	return filtered
}

// sortRows sorts rows: problematic first, then by region and name
func sortRows(rows []row) {
	sort.Slice(rows, func(i, j int) bool {
		pi := isProblematic(rows[i])
		pj := isProblematic(rows[j])
		if pi != pj {
			return pi // problematic first
		}
		if rows[i].Region == rows[j].Region {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].Region < rows[j].Region
	})
}

// countProblems counts various problem types in the rows
type problemCounts struct {
	problematicCount      int
	alarmStateCount       int
	insufficientDataCount int
	disabledCount         int
	noActionsCount        int
}

func countProblems(rows []row) problemCounts {
	var counts problemCounts
	for _, r := range rows {
		if isProblematic(r) {
			counts.problematicCount++
		}
		if r.State == "ALARM" {
			counts.alarmStateCount++
		}
		if r.State == "INSUFFICIENT_DATA" {
			counts.insufficientDataCount++
		}
		if !r.ActionsEnabled {
			counts.disabledCount++
		}
		if r.AlarmActions == 0 && r.OKActions == 0 && r.InsufActions == 0 {
			counts.noActionsCount++
		}
	}
	return counts
}

// getFilterReasons returns a list of active filter reasons for error messages
func getFilterReasons(opts filterOptions) []string {
	var reasons []string
	if opts.onlyBroken {
		reasons = append(reasons, "--only-broken")
	}
	if opts.stateFilter != "" {
		reasons = append(reasons, fmt.Sprintf("--state=%s", opts.stateFilter))
	}
	if opts.noActions {
		reasons = append(reasons, "--no-actions")
	}
	if opts.actionsDisabled {
		reasons = append(reasons, "--actions-disabled")
	}
	if opts.stale > 0 {
		reasons = append(reasons, fmt.Sprintf("--stale=%v", opts.stale))
	}
	return reasons
}
