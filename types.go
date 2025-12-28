package main

import "time"

// row represents a single CloudWatch alarm row
type row struct {
	Region                string
	Name                  string
	State                 string
	ActionsEnabled        bool
	AlarmActions          int
	OKActions             int
	InsufActions          int
	StateUpdatedTimestamp *time.Time
}

// alarmJSON represents an alarm in JSON output format
type alarmJSON struct {
	Region         string `json:"region"`
	Name           string `json:"name"`
	State          string `json:"state"`
	Enabled        bool   `json:"enabled"`
	AlarmActions   int    `json:"alarm_actions"`
	OKActions      int    `json:"ok_actions"`
	InsufActions   int    `json:"insufficient_actions"`
	LastChanged    string `json:"last_changed"`
	LastChangedISO string `json:"last_changed_iso,omitempty"`
}

// outputJSON represents the complete JSON output structure
type outputJSON struct {
	Alarms  []alarmJSON `json:"alarms"`
	Summary struct {
		TotalAlarms           int `json:"total_alarms"`
		AlarmsShown           int `json:"alarms_shown"`
		ProblematicAlarms     int `json:"problematic_alarms,omitempty"`
		InAlarmState          int `json:"in_alarm_state"`
		InsufficientDataState int `json:"insufficient_data_state"`
		ActionsDisabled       int `json:"actions_disabled"`
		NoActionsConfigured   int `json:"no_actions_configured"`
	} `json:"summary"`
}
