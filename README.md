# AWS CloudWatch Alarm Audit CLI

Small, read-only CLI tool to audit **CloudWatch alarms across AWS regions**.

It helps identify alarms that:

* have **no alarm actions configured**
* have **alarm actions disabled**
* stay in **ALARM** or **INSUFFICIENT_DATA** state for a long time

---

## How to run

```bash
go run .
```

---

## Common usage

```bash
# Broken alarms across all regions (no actions, disabled actions, stale states >7d)
go run . --only-broken --stale 7d
```

```bash
# Alarms currently in ALARM state (JSON output, specific region)
go run . --regions eu-west-1 --state ALARM --json
```

```bash
# Alarms that do nothing when they fire
go run . --no-actions --actions-disabled
```

---

## Output

By default, the tool prints a table with:

* region
* alarm name
* current state
* time since last state change
* detected issues

Use `--json` for structured output.

---

## Required AWS permissions

* `cloudwatch:DescribeAlarms`
* `ec2:DescribeRegions`

---

## Scope

* read-only
* no alarm changes
* no auto-fixing
* no dashboards

## Feedback

If you've run into similar issues or have ideas for improvements, feel free to open an issue or reach out.

Email: wojtekrybakiewicz@proton.me
