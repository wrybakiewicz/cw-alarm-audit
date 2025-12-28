package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// printTable outputs the results in a formatted table
func printTable(rows []row, totalCount int, opts filterOptions, fullNames bool) {
	// Apply filters
	filtered := applyFilters(rows, opts)
	rows = filtered

	if len(rows) == 0 {
		reasons := getFilterReasons(opts)
		if len(reasons) > 0 {
			fmt.Printf("No alarms found matching filters: %s\n", strings.Join(reasons, ", "))
		} else {
			fmt.Println("No alarms found")
		}
		return
	}

	// Sort rows
	sortRows(rows)

	// Count problems for summary
	counts := countProblems(rows)

	// Create table with go-pretty
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleColoredBright)
	t.Style().Options.SeparateRows = false
	t.Style().Options.DrawBorder = true
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateHeader = true
	t.Style().Color.Header = []text.Color{text.BgHiBlue, text.FgHiWhite}
	t.Style().Color.Row = []text.Color{text.FgHiWhite}
	t.Style().Color.RowAlternate = []text.Color{text.FgWhite}

	// Configure columns with descriptive headers
	t.AppendHeader(table.Row{
		"REGION",
		"ALARM_NAME",
		"STATE",
		"ENABLED",
		"ALARM ACTIONS",
		"OK ACTIONS",
		"INSUFFICIENT ACTIONS",
		"LAST CHANGED",
	})

	// Calculate name column width
	nameWidth := 60
	if fullNames {
		maxLen := 0
		for _, r := range rows {
			if len(r.Name) > maxLen {
				maxLen = len(r.Name)
			}
		}
		if maxLen > 60 {
			nameWidth = maxLen + 10
		}
	}

	// Configure column alignments and widths
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignLeft, WidthMax: 15},
		{Number: 2, Align: text.AlignLeft, WidthMax: nameWidth, WidthMin: nameWidth},
		{Number: 3, Align: text.AlignLeft, WidthMax: 18},
		{Number: 4, Align: text.AlignCenter, WidthMax: 8},
		{Number: 5, Align: text.AlignRight, WidthMax: 14}, // ALARM ACTIONS
		{Number: 6, Align: text.AlignRight, WidthMax: 11}, // OK ACTIONS
		{Number: 7, Align: text.AlignRight, WidthMax: 20}, // INSUFFICIENT ACTIONS
		{Number: 8, Align: text.AlignLeft, WidthMax: 13},  // LAST CHANGED
	})

	// Add rows with color coding for problematic alarms
	now := time.Now()
	for _, r := range rows {
		enabled := "YES"
		if !r.ActionsEnabled {
			enabled = "NO"
		}
		alarmName := r.Name
		if !fullNames && len(r.Name) > nameWidth {
			alarmName = truncate(r.Name, nameWidth)
		}

		// Color code based on state and problems
		state := r.State
		if r.State == "ALARM" {
			state = text.FgRed.Sprint(r.State)
		} else if r.State == "INSUFFICIENT_DATA" {
			state = text.FgYellow.Sprint(r.State)
		}

		enabledText := enabled
		if !r.ActionsEnabled {
			enabledText = text.FgRed.Sprint(enabled)
		}

		alarmActions := fmt.Sprintf("%d", r.AlarmActions)
		if r.AlarmActions == 0 && r.State == "ALARM" {
			alarmActions = text.FgRed.Sprint(alarmActions)
		}

		// Format last changed time
		lastChanged := formatLastChanged(r.StateUpdatedTimestamp, now)
		if r.StateUpdatedTimestamp != nil {
			timeSince := now.Sub(*r.StateUpdatedTimestamp)
			// Color code if stale
			if opts.stale > 0 && timeSince >= opts.stale {
				lastChanged = text.FgRed.Sprint(lastChanged)
			} else if timeSince >= 7*24*time.Hour {
				lastChanged = text.FgYellow.Sprint(lastChanged)
			}
		}

		t.AppendRow(table.Row{
			r.Region,
			alarmName,
			state,
			enabledText,
			alarmActions,
			fmt.Sprintf("%d", r.OKActions),
			fmt.Sprintf("%d", r.InsufActions),
			lastChanged,
		})
	}

	// Render table
	t.Render()

	// Summary
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Total alarms: %d\n", totalCount)
	fmt.Printf("  Alarms shown (after filters): %d\n", len(rows))
	if !opts.onlyBroken {
		fmt.Printf("  Problematic alarms: %d\n", counts.problematicCount)
	}
	fmt.Printf("  In ALARM state: %d\n", counts.alarmStateCount)
	fmt.Printf("  INSUFFICIENT_DATA state: %d\n", counts.insufficientDataCount)
	fmt.Printf("  Actions disabled: %d\n", counts.disabledCount)
	fmt.Printf("  No actions configured: %d\n", counts.noActionsCount)
}
