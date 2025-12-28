package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// printJSON outputs the results in JSON format
func printJSON(rows []row, totalCount int, opts filterOptions) {
	// Apply filters
	filtered := applyFilters(rows, opts)
	rows = filtered

	// Count problems for summary
	counts := countProblems(rows)

	// Build JSON output
	now := time.Now()
	var alarms []alarmJSON
	for _, r := range rows {
		lastChanged := formatLastChanged(r.StateUpdatedTimestamp, now)
		lastChangedISO := ""
		if r.StateUpdatedTimestamp != nil {
			lastChangedISO = r.StateUpdatedTimestamp.Format(time.RFC3339)
		}

		alarms = append(alarms, alarmJSON{
			Region:         r.Region,
			Name:           r.Name,
			State:          r.State,
			Enabled:        r.ActionsEnabled,
			AlarmActions:   r.AlarmActions,
			OKActions:      r.OKActions,
			InsufActions:   r.InsufActions,
			LastChanged:    lastChanged,
			LastChangedISO: lastChangedISO,
		})
	}

	output := outputJSON{
		Alarms: alarms,
		Summary: struct {
			TotalAlarms           int `json:"total_alarms"`
			AlarmsShown           int `json:"alarms_shown"`
			ProblematicAlarms     int `json:"problematic_alarms,omitempty"`
			InAlarmState          int `json:"in_alarm_state"`
			InsufficientDataState int `json:"insufficient_data_state"`
			ActionsDisabled       int `json:"actions_disabled"`
			NoActionsConfigured   int `json:"no_actions_configured"`
		}{
			TotalAlarms:           totalCount,
			AlarmsShown:           len(rows),
			ProblematicAlarms:     counts.problematicCount,
			InAlarmState:          counts.alarmStateCount,
			InsufficientDataState: counts.insufficientDataCount,
			ActionsDisabled:       counts.disabledCount,
			NoActionsConfigured:   counts.noActionsCount,
		},
	}

	// Only include problematic_alarms if not using --only-broken filter
	if opts.onlyBroken {
		output.Summary.ProblematicAlarms = 0
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
