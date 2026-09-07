---
page_title: "joopler_connector Resource - joopler"
subcategory: ""
description: |-
  A connector's non-secret configuration.
---

# joopler_connector (Resource)

A connector's **non-secret** configuration.

~> **Never put secrets in `config`.** Credentials are set separately and are
stored in AWS Secrets Manager, never in the Joopler database and never returned
by the API. Anything you put in `config` lands in Terraform state in plain text,
which is exactly what keeping secrets out of this resource avoids.

## Example usage

```terraform
resource "joopler_connector" "aws" {
  key     = "aws"
  enabled = true
  config = {
    region = "us-east-2"
  }
}
```

## Schema

### Required

- `key` (String) Connector key (e.g. `aws`, `okta`, `github`).

### Optional

- `enabled` (Boolean) Whether the connector is enabled. Defaults to `true`.
- `config` (Map of String) Non-secret config (e.g. `region`, `orgUrl`).
