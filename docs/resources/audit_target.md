---
page_title: "joopler_audit_target Resource - joopler"
subcategory: ""
description: |-
  The tenant's audit-readiness target date (singleton).
---

# joopler_audit_target (Resource)

The tenant's audit-readiness target date. Drives the countdown and the journey
nudges. This is a **singleton** - one target per workspace, so declare it once.

## Example usage

```terraform
resource "joopler_audit_target" "soc2" {
  target_date = "2027-01-31"
}
```

## Schema

### Required

- `target_date` (String) Target date, ISO `YYYY-MM-DD`.
