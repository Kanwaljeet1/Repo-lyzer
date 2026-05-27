## Description

Fixes #285

The scheduler daemon in Repo-lyzer had two major calculation flaws which are resolved in this PR:

1. **Incorrect Execution Time Duration:** In `internal/scheduler/scheduler.go`, the job duration during analysis execution was evaluated as `time.Since(time.Now())`. Because `time.Now()` was evaluated instantly during the assignment, this always produced a duration close to 0s, failing to capture the actual time spent running the analysis.
2. **Mock Next-Run Calculation:** Both `calculateNextRunTime` (in `scheduler.go`) and `calculateNextRun` (in `settings.go`) were previously hardcoded to always add 24 hours (`time.Now().Add(24 * time.Hour)`), completely ignoring custom cron interval settings and resulting in scheduler drift and incorrect next-execution display metrics.

## Changes Made
- **Tracked `startTime`:** Introduced `startTime := time.Now()` at the beginning of `executeJob()` and updated the `Duration` field in `compactCfg` to use `time.Since(startTime)`, correctly measuring the exact analysis execution duration.
- **Implemented `robfig/cron/v3` parser:** Updated both `calculateNextRunTime` and `calculateNextRun` to accurately calculate the next occurrence based on the configured custom cron expression. They now use `cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)` to accurately generate a `time.Time` for the next scheduled occurrence based on interval schedules (weekly, monthly, custom, etc.).

## GSSoC 2026
This PR is submitted as part of GSSoC 2026. Closes #285.
