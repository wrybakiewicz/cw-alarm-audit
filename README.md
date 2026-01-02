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

---

## What this tool does

It identifies CloudWatch alarms that:

- have no alarm actions configured
- have alarm actions disabled
- stay in **ALARM** or **INSUFFICIENT_DATA** state longer than a given threshold

The output is intended as a **starting point for human review**.

---

## What this tool does NOT do

- It does not decide if an alarm is correct
- It does not modify alarms
- It does not solve alert noise or threshold tuning

If many alarms are flagged, the problem is usually process-related (e.g. ownership or reviews), not CloudWatch itself.

---

## Installation

Prebuilt binaries are available on the GitHub releases page:

https://github.com/wrybakiewicz/cw-alarm-audit/releases

Example (macOS, Apple Silicon):

```bash
curl -L https://github.com/wrybakiewicz/cw-alarm-audit/releases/download/v0.1.0/cw-alarm-audit_0.1.0_darwin_arm64.tar.gz -o cw-alarm-audit.tar.gz
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

---

## Example output

```
REGION     | ALARM_NAME       | STATE   | ENABLED | ALARM_ACTIONS | OK_ACTIONS | INSUFFICIENT_ACTIONS | LAST_CHANGED
eu-west-1  | api-5xx-errors   | ALARM   | false   | 0             | 0          | 0                    | 12d ago
us-east-1  | db-cpu-high      | OK      | true    | 0             | 0          | 0                    | 45d ago
```

In this example:
- `api-5xx-errors` is disabled and has no actions configured
- `db-cpu-high` is enabled but does not notify anyone

---

## Required AWS permissions

The tool only requires read-only access:

- `cloudwatch:DescribeAlarms`
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
