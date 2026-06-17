<p align="center">
  <img src="assets/logo.svg" alt="Terrasentry logo" width="540">
</p>

# Terrasentry

**How locked-in to one cloud is your Terraform? Terrasentry gives you the answer as a single score.**

Terrasentry reads your Terraform plan and rates it from **0.00** (everything is
tied to one cloud provider and would be painful to move) to **1.00** (built from
portable, swappable pieces). It scores every resource, every module, and the
whole project — so you can finally answer the question leadership keeps asking:
*how hard would it actually be to leave AWS, Azure, or GCP?*

It is one small open-source program (a single file, MIT-licensed). It runs
entirely on your own machine — no internet connection, no account, no API key.

> **Status: early MVP.** The portability score and the `scan` command work
> today. Security checks, cost checks, naming-convention checks, the CI gate,
> and the editor-integration server are planned — see the [roadmap](#roadmap).

## What you get today

- **One number for cloud lock-in** — run it on a project and see at a glance how
  portable your infrastructure is.
- **The worst offenders, named** — it points to the exact resources dragging the
  score down (proprietary login services, serverless lock-in, and so on).
- **A simple PASS / FAIL** — set a minimum score, and the scan tells you whether
  the project clears it, ready to plug into code review later.
- **Fully offline** — nothing ever leaves your machine.

---

## Why this exists

Existing Terraform governance tools each cover a slice and skew enterprise:

| Tool | What it does | What it misses |
|------|--------------|----------------|
| **Firefly** | discovery, codegen, drift, AI remediation | expensive, enterprise-only |
| **Infracost** | excellent AI-native cost checks | cost *only* |
| **Checkov / Trivy** | security misconfig scanning | flags but doesn't fix; not AI-native |
| **Spacelift / env0 / Scalr** | run orchestration + policy-as-code | you hand-write OPA/Sentinel; enterprise pricing |
| **TerraGuard MCP** | MCP guardrail + 23 rules | BSL-licensed (not open); no cost; **no portability score**; fixes are guidance, not diffs |
| **HashiCorp Terraform MCP** | approved-module discovery for AI | not policy enforcement; no portability score |

**Nobody scores how cloud-locked your infrastructure is.** That is
Terrasentry's wedge. Leadership keeps asking "how portable is our infra?" and
there has been no tool that answers it. We do, per resource, per module, and
per repo — with a trend you can gate on in CI.

The second differentiator: a **truly open core (MIT)**, affordable for
mid-size teams, that unifies four dimensions instead of making you stitch
together four tools and a pile of OPA policy.

## The four policy dimensions

1. **Portability** *(live)* — every resource type is scored `0.0` (fully
   cloud-locked, e.g. `aws_cognito_user_pool`) to `1.0` (portable commodity,
   e.g. `aws_instance`). Aggregated into a module- and repo-level lock-in
   score you can threshold in CI. **This is the unique feature.**
2. **Security** *(roadmap)* — misconfig checks (public buckets, open security
   groups, unencrypted volumes); may wrap Trivy/Checkov when present.
3. **Cost** *(roadmap)* — flag obviously expensive resources via heuristics,
   no live pricing API needed.
4. **Conventions** *(roadmap)* — naming patterns, required tags, approved
   module sources — all from `terrasentry.yaml`.

For every violation the roadmap goal is a **suggested HCL fix (a diff)**, not
just a flag — directly answering the "auto-fix in the PR" gap.

## Install & build

Requires Go 1.25+.

```bash
go build -o terrasentry ./cmd/terrasentry
```

This produces a single static binary with the portability dataset embedded.

## Usage

```bash
# 1. Produce a plan JSON from your Terraform project:
terraform plan -out tfplan
terraform show -json tfplan > plan.json

# 2. Scan it:
terrasentry scan --plan plan.json                 # human-readable report
terrasentry scan --plan plan.json --json          # machine-readable JSON
terrasentry scan --plan plan.json --config terrasentry.yaml
```

### Try it now against the bundled sample

```bash
terrasentry scan --plan examples/sample-project/plan.json --config terrasentry.yaml
```

Produces a repo lock-in score of `0.48` (below the `0.60` policy minimum),
flags Cognito / DynamoDB / API Gateway / Lambda as the most locked-in
resources, and reports `FAIL`.

## Configuration (`terrasentry.yaml`)

The file is **optional** — delete it and built-in defaults apply.

```yaml
version: 1

portability:
  min_score: 0.6    # fail the repo below this aggregate lock-in score
  warn_below: 0.4   # flag any single resource more locked-in than this
  # dataset_path: ./overrides.yaml   # optional: bring your own scores

budget:                              # roadmap
  monthly_limit_usd: 1000
conventions:                         # roadmap
  required_tags: [Environment, Owner, CostCenter]
  naming_pattern: "^[a-z][a-z0-9-]+$"
  approved_module_sources:
    - "registry.terraform.io/terraform-aws-modules/*"
security:                            # roadmap
  enabled: true
```

## The portability dataset

Scores live in [`data/portability.yaml`](data/portability.yaml), embedded in
the binary and covering ~65 common AWS/Azure/GCP resource types. Each entry
carries a documented reason:

```yaml
resources:
  aws_s3_bucket:         { score: 0.80, reason: "S3 API is a de-facto standard" }
  aws_cognito_user_pool: { score: 0.05, reason: "Auth fully proprietary" }
```

Scores are deliberate heuristics, not absolute truth. Override per-org via
`portability.dataset_path`. Unknown types score a neutral `0.5` and are listed
in the report so you know what's uncovered.

## Architecture

```
cmd/         CLI entry + scan command (stdlib flag, no framework)
parser/      terraform show -json  ->  normalized, flat resource list
rules/       pluggable Rule engine + Finding model
  portability/   the wedge: scorer, dataset loader, Report, engine Rule
config/      terrasentry.yaml loader (defaults when absent)
report/      terminal + JSON renderers
data/        embedded seed portability dataset
examples/    sample Terraform project + committed plan.json
```

Design decisions:

- **Pluggable rule engine.** Each dimension implements a small `Rule`
  interface; the engine just runs registered rules. New rules and resource
  mappings are data/file changes, not engine changes. Security/cost/
  conventions ship today as registered stubs so the wiring is real.
- **Offline by default.** The scan makes zero network calls — a feature
  competitors gate behind paid tiers. The dataset is embedded via `go:embed`.
- **Plan-JSON first.** We consume `terraform show -json`, the stable
  documented contract, rather than parsing raw HCL (raw `.tf` parsing is a
  later add).
- **Single binary, MIT core**, open-core model.
- **Free, and bring-your-own-key.** Terrasentry is free and needs no API key to
  run. Any future AI features (suggested fixes, the MCP server) use a key *you*
  provide through an environment variable — the tool never bundles a key or
  routes requests through anyone else's account.

## Roadmap

- [ ] `terrasentry ci` — PR-comment markdown + non-zero exit on threshold
      breach (budget-style gate).
- [ ] **MCP server** (`terrasentry mcp`) — expose `check_policy`,
      `score_portability`, `suggest_fix` as MCP tools so AI coding assistants
      (e.g. Cursor) enforce policy *during* AI code generation.
- [ ] Security rules (public buckets, open SGs, unencrypted volumes; wrap
      Trivy/Checkov).
- [ ] Cost heuristics.
- [ ] Conventions: required tags, naming, approved modules.
- [ ] Suggested HCL fixes as diffs.
- [ ] Raw `.tf`/HCL parsing alongside plan JSON.
- [ ] Portability trend over time.
- [ ] Test suite.

## License

MIT (core).
