# terraform-provider-joopler

Manage the **scaffolding** of a Joopler compliance program as code: control and
policy ownership, vendors/subprocessors, the audit target, and connectors.

**By design, this provider cannot sign off on policies or attest evidence.**
Those are human acts. The provider drives only the versioned write surface in
[`docs/api-v1.md`](../docs/api-v1.md); the API default-denies everything else to
a key. So Terraform owns the structure of your program and a human owns the
claims - both auditable, the structure as code and the sign-offs as signed ledger
records.

## Usage

```hcl
terraform {
  required_providers {
    joopler = {
      source = "joopler/joopler"
    }
  }
}

provider "joopler" {
  # api_url defaults to https://api.joopler.com
  api_key = var.joopler_api_key # a WRITE-scoped key (jpl_...), or set JOOPLER_API_KEY
}

resource "joopler_control_owner" "cloudtrail" {
  control_key = "aws-cloudtrail-enabled"
  owner_email = "security@acme.com"
}

resource "joopler_policy_owner" "aup" {
  policy_key = "acceptable-use"
  owner      = "legal@acme.com"
}

resource "joopler_vendor" "stripe" {
  name        = "Stripe"
  category    = "Payments"
  risk_tier   = "high"
  data_access = "pii"
  owner_email = "security@acme.com"
}

resource "joopler_audit_target" "soc2" {
  target_date = "2027-01-31"
}

resource "joopler_connector" "aws" {
  key     = "aws"
  enabled = true
  config  = { region = "us-east-2" }
}

data "joopler_environment" "this" {}
data "joopler_control_status" "this" {}

output "readiness" {
  value = "${data.joopler_control_status.this.passing}/${data.joopler_control_status.this.total} controls passing"
}
```

## Resources

| Resource | Manages |
| --- | --- |
| `joopler_control_owner` | The accountable owner of a control |
| `joopler_policy_owner` | The owner of a policy (sign-off stays human) |
| `joopler_vendor` | A vendor / subprocessor (delete = offboard) |
| `joopler_audit_target` | The audit-readiness target date (singleton) |
| `joopler_connector` | A connector's non-secret config (secrets set separately) |

## Data sources

| Data source | Returns |
| --- | --- |
| `joopler_control_status` | Per-control status + rollup counts |
| `joopler_environment` | Connected systems, IdP, endpoints, subprocessors, workforce |

## Building and publishing

This repository is a standalone Go module intended to live at
`github.com/joopler/terraform-provider-joopler`.

```
go mod tidy
go build ./...
go test ./...
```

## Publishing to the public Terraform Registry

The release is automated by `.github/workflows/release.yml` (GoReleaser). One
time:

1. Create a GPG signing key and export it (see the repo's GPG setup notes).
2. Create the public repo **`joopler/terraform-provider-joopler`** and push the
   contents of this `provider/` directory to its root (so `go.mod`,
   `.goreleaser.yml`, `terraform-registry-manifest.json`, and `.github/` sit at
   the repo root).
3. Add two repo secrets: `GPG_PRIVATE_KEY` (armored private key) and `PASSPHRASE`
   (omit if the key has none). The workflow derives `GPG_FINGERPRINT` from the
   imported key, so you do not set it manually.
4. Add the **public** GPG key to registry.terraform.io under your publisher's
   **Settings -> GPG Keys**.
5. Connect the repo in the Terraform Registry (Publish -> Provider) under the
   `joopler` namespace.

Then every release is just a tag:

```
git tag v0.1.0 && git push origin v0.1.0
```

The workflow builds the signed cross-platform archives and publishes the GitHub
release the Registry ingests.

Note: this provider is developed inside the trustcenter monorepo under
`provider/`; the publish repo is a subtree/copy of that directory, which is why
the release workflow lives at `provider/.github/workflows/release.yml` here.
