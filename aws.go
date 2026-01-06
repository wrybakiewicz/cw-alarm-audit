package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// resolveRegions resolves the list of regions to scan
// If regionsCSV is empty, it returns all opted-in regions
func resolveRegions(ctx context.Context, cfg aws.Config, regionsCSV string) ([]string, error) {
	if strings.TrimSpace(regionsCSV) != "" {
		parts := strings.Split(regionsCSV, ",")
		var rs []string
		for _, p := range parts {
			r := strings.TrimSpace(p)
			if r != "" {
				rs = append(rs, r)
			}
		}
		if len(rs) == 0 {
			return nil, fmt.Errorf("no valid regions in --regions")
		}
		return rs, nil
	}

	// Default: scan all regions. Use EC2 DescribeRegions.
	// Use a "home" region for this call; if cfg.Region empty, default to us-east-1.
	homeCfg := cfg
	if homeCfg.Region == "" {
		homeCfg.Region = "us-east-1"
	}
	ec2c := ec2.NewFromConfig(homeCfg)
	resp, err := ec2c.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	var rs []string
	for _, rr := range resp.Regions {
		if rr.RegionName == nil || *rr.RegionName == "" {
			continue
		}
		// Filter out regions your account hasn't opted into
		// Valid values commonly include: "opted-in", "opt-in-not-required", "not-opted-in"
		if rr.OptInStatus != nil {
			s := string(*rr.OptInStatus)
			if s != "opted-in" && s != "opt-in-not-required" {
				continue
			}
		}
		rs = append(rs, *rr.RegionName)
	}

	if len(rs) == 0 {
		return nil, fmt.Errorf("DescribeRegions returned 0 regions")
	}
	return rs, nil
}

// getAlarmStateFlapsCount retrieves the number of state changes for an alarm within a time window
func getAlarmStateFlapsCount(ctx context.Context, cwc *cloudwatch.Client, alarmName string, window time.Duration) (int, error) {
	now := time.Now()
	startTime := now.Add(-window)

	var totalFlaps int
	var nextToken *string

	for {
		resp, err := cwc.DescribeAlarmHistory(ctx, &cloudwatch.DescribeAlarmHistoryInput{
			AlarmName:       aws.String(alarmName),
			StartDate:       aws.Time(startTime),
			EndDate:         aws.Time(now),
			HistoryItemType: types.HistoryItemTypeStateUpdate,
			NextToken:       nextToken,
			MaxRecords:      aws.Int32(100),
		})
		if err != nil {
			return 0, err
		}

		totalFlaps += len(resp.AlarmHistoryItems)

		nextToken = resp.NextToken
		if nextToken == nil || *nextToken == "" {
			break
		}
	}

	return totalFlaps, nil
}

// scanRegion scans CloudWatch alarms in a specific region
func scanRegion(ctx context.Context, cfg aws.Config, region, namePrefix string) ([]row, error) {
	rcfg := cfg
	rcfg.Region = region

	cwc := cloudwatch.NewFromConfig(rcfg)

	var out []row
	var nextToken *string

	for {
		in := &cloudwatch.DescribeAlarmsInput{
			NextToken:  nextToken,
			MaxRecords: aws.Int32(100),
		}
		if namePrefix != "" {
			in.AlarmNamePrefix = aws.String(namePrefix)
		}

		resp, err := cwc.DescribeAlarms(ctx, in)
		if err != nil {
			return nil, err
		}

		for _, a := range resp.MetricAlarms {
			var stateUpdated *time.Time
			if a.StateUpdatedTimestamp != nil {
				stateUpdated = a.StateUpdatedTimestamp
			}
			out = append(out, row{
				Region:                region,
				Name:                  aws.ToString(a.AlarmName),
				State:                 string(a.StateValue),
				ActionsEnabled:        a.ActionsEnabled != nil && *a.ActionsEnabled,
				AlarmActions:          len(a.AlarmActions),
				OKActions:             len(a.OKActions),
				InsufActions:          len(a.InsufficientDataActions),
				StateUpdatedTimestamp: stateUpdated,
				StateFlapsCount:       0, // Will be populated later if noisy mode is enabled
			})
		}

		nextToken = resp.NextToken
		if nextToken == nil || *nextToken == "" {
			break
		}
	}
	return out, nil
}

// enrichRowsWithFlapsCount enriches rows with state flaps count for noisy detection
func enrichRowsWithFlapsCount(ctx context.Context, cfg aws.Config, rows []row, window time.Duration) {
	// Group rows by region to batch API calls
	rowsByRegion := make(map[string][]*row)
	for i := range rows {
		region := rows[i].Region
		rowsByRegion[region] = append(rowsByRegion[region], &rows[i])
	}

	// Process each region
	for region, regionRows := range rowsByRegion {
		rcfg := cfg
		rcfg.Region = region
		cwc := cloudwatch.NewFromConfig(rcfg)

		// Get flaps count for each alarm in this region
		for _, r := range regionRows {
			flaps, err := getAlarmStateFlapsCount(ctx, cwc, r.Name, window)
			if err != nil {
				// If we can't get history, leave flaps count at 0
				continue
			}
			r.StateFlapsCount = flaps
		}
	}
}
