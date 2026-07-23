# Example: a slice of a Joopler program as code. Ownership, a vendor, the audit
# target, and a connector - the scaffolding. Sign-off stays a human act in the app.

terraform {
  required_providers {
    joopler = {
      source = "joopler/joopler"
    }
  }
}

variable "joopler_api_key" {
  type      = string
  sensitive = true
}

provider "joopler" {
  api_key = var.joopler_api_key
}

locals {
  control_owners = {
    "aws-cloudtrail-enabled"      = "security@acme.com"
    "aws-config-recorder"         = "platform@acme.com"
    "security-awareness-training" = "people@acme.com"
  }
}

resource "joopler_control_owner" "this" {
  for_each    = local.control_owners
  control_key = each.key
  owner_email = each.value
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

data "joopler_control_status" "this" {}

output "readiness" {
  value = "${data.joopler_control_status.this.passing}/${data.joopler_control_status.this.total} controls passing"
}
