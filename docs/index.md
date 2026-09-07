---
page_title: "Joopler Provider"
subcategory: ""
description: |-
  Manage the scaffolding of a Joopler compliance program as code. Sign-off and evidence attestation are human acts and are intentionally not managed by this provider.
---

# Joopler Provider

Manage the **scaffolding** of a Joopler compliance program as code: control and
policy ownership, vendors and subprocessors, the audit target, and connector
configuration.

**This provider cannot sign off on policies or attest evidence.** Those are human
acts. It drives only the versioned write surface of the Joopler API, which
default-denies everything else to a key. So Terraform owns the structure of your
program and a person owns the claims: the structure reviewable as code, the
sign-offs as signed ledger records.

## Before you start: you need a workspace

This provider configures an **existing** Joopler workspace. It does not create an
account, so `terraform init` will succeed for anyone and `terraform plan` will
stop with `Missing Joopler API key` until you supply one.

- **Have a workspace?** Mint a write-scoped key (`jpl_...`) at
  [app.joopler.com/developers](https://app.joopler.com/developers).
- **Do not have one yet?** Joopler is currently invitation-only. Get in touch at
  [joopler.com](https://joopler.com).

The key is write-scoped, and the API default-denies anything outside the surface
these resources cover. A leaked key cannot read your evidence or sign anything.

## Example usage

```terraform
terraform {
  required_providers {
    joopler = {
      source = "joopler/joopler"
    }
  }
}

provider "joopler" {
  # api_key defaults to the JOOPLER_API_KEY environment variable,
  # which is the better place for it - keep keys out of state and VCS.
  api_key = var.joopler_api_key
}

resource "joopler_control_owner" "cloudtrail" {
  control_key = "aws-cloudtrail-enabled"
  owner_email = "security@acme.com"
}

resource "joopler_audit_target" "soc2" {
  target_date = "2027-01-31"
}

data "joopler_control_status" "this" {}

output "readiness" {
  value = "${data.joopler_control_status.this.passing}/${data.joopler_control_status.this.total} controls passing"
}
```

## Schema

### Optional

- `api_key` (String, Sensitive) Write-scoped tenant API key (`jpl_...`). Defaults
  to the `JOOPLER_API_KEY` environment variable. Mint one at
  [app.joopler.com/developers](https://app.joopler.com/developers); it requires an
  existing Joopler workspace.
- `api_url` (String) Joopler API base URL. Defaults to `https://api.joopler.com`,
  or the `JOOPLER_API_URL` environment variable.

## What this provider deliberately does not do

Policy sign-off and control attestation are absent on purpose, not pending. Both
produce a signed, timestamped document written to a tamper-evident ledger, and
both assert that a person reviewed something and stands behind it. A `terraform
apply` that could manufacture those would make the strongest evidence in the
system the easiest thing in it to fake.
