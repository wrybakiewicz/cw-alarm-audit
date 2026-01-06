# AWS CloudWatch Alarm Audit CLI

A small, read-only CLI tool to audit **CloudWatch alarms across AWS regions**.

It helps find CloudWatch alarms that no longer behave as expected,
for example alarms with no actions or alarms stuck in a bad state.

---

## Why this exists

In larger AWS setups, CloudWatch alarms tend to pile up over time.

You end up with things like:
- alarms with no actions
- alarms that were disabled and never revisited
- alarms stuck in `ALARM` or `INSUFFICIENT_DATA` for days or weeks

After a while, people stop trusting alerts.
They get ignored, silenced, or left broken.

This tool helps you find those alarms.

In practice, these issues often surface during alerting reviews or incident retrospectives, not during initial setup.

---

## What this tool does

It identifies CloudWatch alarms that:

- have no alarm actions configured
- have alarm actions disabled
- stay in **ALARM** or **INSUFFICIENT_DATA** state longer than a given threshold
- frequently change state (noisy/flapping alarms) within a time window

The output is intended as a **starting point for human review**.

---

## What this tool does NOT do

- It does not decide if an alarm is correct
- It does not modify alarms
- It does not fix alert noise or perform threshold tuning

If many alarms are flagged, the problem is usually process-related (e.g. ownership or reviews), not CloudWatch itself.

---

## Installation

Prebuilt binaries are available on the GitHub releases page:

https://github.com/wrybakiewicz/cw-alarm-audit/releases

Example (macOS, Apple Silicon):

```bash
curl -L https://github.com/wrybakiewicz/cw-alarm-audit/releases/download/v1.1.0/cw-alarm-audit_1.1.0_darwin_arm64.tar.gz -o cw-alarm-audit.tar.gz
tar -xzf cw-alarm-audit.tar.gz
chmod +x cw-alarm-audit
./cw-alarm-audit
```

---

## Quick start

Scan for commonly broken alarms across all regions:

```bash
./cw-alarm-audit --only-broken --stale 7d
```

---

## Common usage

```bash
# Broken alarms across all regions (no actions, disabled actions, stale states >7d)
./cw-alarm-audit --only-broken --stale 7d
```

```bash
# Alarms currently in ALARM state (JSON output, specific region)
./cw-alarm-audit --regions eu-west-1 --state ALARM --json
```

```bash
# Alarms that do nothing when they fire
./cw-alarm-audit --no-actions --actions-disabled
```

```bash
# Alarms that frequently change state (noisy/flapping alarms)
# Shows alarms with at least 5 state changes in the last 24 hours
./cw-alarm-audit --noisy
```

```bash
# Custom noisy alarm detection (10 changes in 48 hours)
./cw-alarm-audit --noisy --noisy-window=48h --noisy-min-flaps=10
```

---

## Example output

Standard output:
```
REGION     | ALARM_NAME       | STATE   | ENABLED | ALARM_ACTIONS | OK_ACTIONS | INSUFFICIENT_ACTIONS | LAST_CHANGED
eu-west-1  | api-5xx-errors   | ALARM   | false   | 0             | 0          | 0                    | 12d ago
us-east-1  | db-cpu-high      | OK      | true    | 0             | 0          | 0                    | 45d ago
```

With `--noisy` flag (shows state flaps count):
```
REGION     | ALARM_NAME       | STATE   | ENABLED | ALARM_ACTIONS | OK_ACTIONS | INSUFFICIENT_ACTIONS | LAST_CHANGED | STATE FLAPS
eu-west-1  | api-5xx-errors   | ALARM   | false   | 0             | 0          | 0                    | 12d ago      | 8
us-east-1  | db-cpu-high      | OK      | true    | 0             | 0          | 0                    | 45d ago      | 12
```

In these examples:
- `api-5xx-errors` is disabled and has no actions configured
- `db-cpu-high` is enabled but does not notify anyone
- Both alarms show their state change count when using `--noisy` flag

---

## Required AWS permissions

The tool only requires read-only access:

- `cloudwatch:DescribeAlarms`
- `cloudwatch:DescribeAlarmHistory` (required when using `--noisy`)
- `ec2:DescribeRegions`

---

## Scope

- uses only AWS read-only APIs
- does not modify or create CloudWatch alarms
- does not write to AWS resources
- produces local output only

---

## Feedback

If you’ve seen similar alerting issues in real environments, or have ideas
for improvements, feedback is very welcome.

You can open an issue or reach out directly:

Email: wojtekrybakiewicz@proton.me
