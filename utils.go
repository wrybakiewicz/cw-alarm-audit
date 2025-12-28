package main

import (
	"fmt"
	"strings"
	"time"
)

// parseDuration parses a duration string, supporting 'd' suffix for days
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := parseFloat(daysStr)
		if err != nil {
			return 0, fmt.Errorf("invalid number of days: %s", daysStr)
		}
		// Convert days to hours
		hours := days * 24
		return time.Duration(hours) * time.Hour, nil
	}
	// Otherwise, use standard time.Duration parser
	return time.ParseDuration(s)
}

// parseFloat parses a float from a string
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// truncate truncates a string to max length, adding "..." if needed
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// formatLastChanged formats the time since last state change as a human-readable string
func formatLastChanged(timestamp *time.Time, now time.Time) string {
	if timestamp == nil {
		return "N/A"
	}
	timeSince := now.Sub(*timestamp)
	if timeSince < time.Hour {
		return fmt.Sprintf("%.0fm ago", timeSince.Minutes())
	} else if timeSince < 24*time.Hour {
		return fmt.Sprintf("%.1fh ago", timeSince.Hours())
	} else {
		days := int(timeSince.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
}
