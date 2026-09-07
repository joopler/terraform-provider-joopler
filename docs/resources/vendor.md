---
page_title: "joopler_vendor Resource - joopler"
subcategory: ""
description: |-
  A vendor / subprocessor in the TPRM register.
---

# joopler_vendor (Resource)

A vendor or subprocessor in the third-party risk register.

~> **Destroy offboards, it does not erase.** Removing this resource marks the
vendor offboarded and Joopler keeps the record. A vendor you used during an audit
period is part of that period's history, so deleting the row outright would
falsify it.

## Example usage

```terraform
resource "joopler_vendor" "stripe" {
  name        = "Stripe"
  category    = "Payments"
  risk_tier   = "high"
  data_access = "pii"
  owner_email = "security@acme.com"
  website     = "https://stripe.com"
}
```

## Schema

### Required

- `name` (String) Vendor name.

### Optional

- `category` (String) Category (e.g. `Payments`).
- `risk_tier` (String) `critical` | `high` | `medium` | `low`.
- `data_access` (String) `none` | `internal` | `confidential` | `pii` | `phi`.
- `owner_email` (String) Vendor owner email.
- `status` (String) `in_review` | `approved` | `offboarded`. Computed when unset.
- `website` (String) Vendor website.
- `notes` (String) Free-text notes.

### Read-Only

- `id` (String) Vendor id.
