# Test projects

Hand-built Terraform scenarios for exercising `terrasentry scan` without a cloud
account. Each folder has a `main.tf` (the Terraform you'd write) and a committed
`plan.json` (what `terraform show -json` would produce). The scanner reads the
`plan.json`.

## Run them all

```bash
./test-projects/run.sh
```

Or one at a time:

```bash
go build -o terrasentry ./cmd/terrasentry
./terrasentry scan --plan test-projects/01-portable-pass/plan.json
```

## Scenarios

| Folder | Stack | Expected repo score | Verdict |
|--------|-------|---------------------|---------|
| `01-portable-pass` | AWS: VM, EBS, S3, VPC, subnet, EKS | ~0.78 | **PASS** |
| `02-locked-fail` | AWS serverless: Cognito, DynamoDB, Lambda, API GW, Step Functions, Kinesis | ~0.14 | **FAIL** (hard) |
| `03-multicloud-unknown` | Azure + GCP + one unscored type | ~0.51 | **FAIL** |
| `04-empty` | No resources (providers/variables only) | n/a | **PASS** |
| `05-data-only` | Data sources only (all skipped) | n/a | **PASS** |

Verdict uses default policy (min_score 0.60, warn_below 0.40) — no config file
passed, so built-in defaults apply.

### What each one proves

- **01-portable-pass** — happy path. Portable primitives only; no resource trips
  the warn threshold; repo clears the minimum. Confirms a clean PASS with zero
  findings.
- **02-locked-fail** — the wedge under load. Six proprietary services, all below
  warn_below, repo far under the minimum. Confirms per-resource findings,
  high/medium severity split, and a hard FAIL.
- **03-multicloud-unknown** — non-AWS scoring (Azure `azurerm_*`, GCP
  `google_*`) plus an unscored type (`aws_braket_quantum_task`). Confirms the
  neutral 0.50 fallback, the "unscored" listing in the report, and that a
  middling mixed estate still fails a 0.60 bar.
- **04-empty** — a plan with no resources. Confirms the "nothing to score" edge:
  repo score is reported as n/a and the result is PASS, not a false 0.00 / FAIL.
- **05-data-only** — a plan with only data sources (lookups). Confirms data
  sources are skipped by the scorer, leaving nothing to score → PASS (n/a).

## Tweak and re-run

- Edit a `plan.json` (swap a locked resource for a portable one) and watch the
  score move.
- Pass `--json` for machine-readable output.
- Pass `--config ../terrasentry.yaml` to apply a stricter/looser policy.
