package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
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
			})
		}

		nextToken = resp.NextToken
		if nextToken == nil || *nextToken == "" {
			break
		}
	}
	return out, nil
}
