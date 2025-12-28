package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	var (
		profile         = flag.String("profile", "", "AWS profile name (optional)")
		regions         = flag.String("regions", "", "Comma-separated list of regions (default: all regions)")
		timeout         = flag.Duration("timeout", 20*time.Second, "API timeout per region")
		namePref        = flag.String("name-prefix", "", "Only alarms with this name prefix (optional)")
		onlyBroken      = flag.Bool("only-broken", false, "Show only broken alarms (ALARM/INSUFFICIENT_DATA state, disabled actions, or no actions)")
		state           = flag.String("state", "", "Filter by alarm state: OK, ALARM, or INSUFFICIENT_DATA")
		noActions       = flag.Bool("no-actions", false, "Show only alarms with no actions configured")
		actionsDisabled = flag.Bool("actions-disabled", false, "Show only alarms with actions disabled")
		staleStr        = flag.String("stale", "", "Show only alarms that haven't changed state since this duration (e.g., 8h, 1d, 7d)")
		jsonOutput      = flag.Bool("json", false, "Output results in JSON format instead of table")
	)

	flag.Parse()

	ctx := context.Background()

	// Load AWS config
	cfgOpts := []func(*config.LoadOptions) error{}
	if *profile != "" {
		cfgOpts = append(cfgOpts, config.WithSharedConfigProfile(*profile))
	}
	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load AWS config: %v\n", err)
		os.Exit(1)
	}

	// Resolve regions - if regions is empty, scan all regions
	targetRegions, err := resolveRegions(ctx, cfg, *regions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve regions: %v\n", err)
		os.Exit(1)
	}
	sort.Strings(targetRegions)

	var out []row
	for _, r := range targetRegions {
		rctx, cancel := context.WithTimeout(ctx, *timeout)
		rows, err := scanRegion(rctx, cfg, r, *namePref)
		cancel()

		if err != nil {
			// Don't fail entire run; just report region error
			fmt.Fprintf(os.Stderr, "region %s: %v\n", r, err)
			continue
		}
		out = append(out, rows...)
	}

	// Validate state filter
	if *state != "" {
		validStates := map[string]bool{"OK": true, "ALARM": true, "INSUFFICIENT_DATA": true}
		if !validStates[*state] {
			fmt.Fprintf(os.Stderr, "Invalid state: %s. Must be one of: OK, ALARM, INSUFFICIENT_DATA\n", *state)
			os.Exit(1)
		}
	}

	// Parse stale duration (supports days with 'd' suffix)
	var staleDuration time.Duration
	if *staleStr != "" {
		var err error
		staleDuration, err = parseDuration(*staleStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid --stale value: %s. Use format like 8h, 1d, 7d, or 24h\n", *staleStr)
			os.Exit(1)
		}
	}

	// Prepare filter options
	opts := filterOptions{
		onlyBroken:      *onlyBroken,
		stateFilter:     *state,
		noActions:       *noActions,
		actionsDisabled: *actionsDisabled,
		stale:           staleDuration,
	}

	// Always show full names by default
	totalCount := len(out)
	if *jsonOutput {
		printJSON(out, totalCount, opts)
	} else {
		printTable(out, totalCount, opts, true)
	}
}
