# Repository audit — 2026-09-20

## Changes

Propagate rule failures, cancellation and filesystem errors instead of reporting clean scans; skip symlinks.

## Validation

go test -race ./... passes.

## Limits

Not a proof that every lint rule detects every issue.
